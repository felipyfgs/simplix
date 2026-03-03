package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportsRepo struct{ db *pgxpool.Pool }

func NewReportsRepo(db *pgxpool.Pool) *ReportsRepo { return &ReportsRepo{db: db} }

type OverviewStats struct {
	TotalContacts      int `json:"total_contacts"`
	OpenConversations  int `json:"open_conversations"`
	ResolvedToday      int `json:"resolved_today"`
	OnlineAgents       int `json:"online_agents"`
}

func (r *ReportsRepo) Overview(ctx context.Context) (*OverviewStats, error) {
	s := &OverviewStats{}
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM contacts`).Scan(&s.TotalContacts)
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM conversations WHERE status='open'`).Scan(&s.OpenConversations)
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM conversations WHERE status='resolved' AND updated_at::date = CURRENT_DATE`).Scan(&s.ResolvedToday)
	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE availability='online'`).Scan(&s.OnlineAgents)
	return s, nil
}

type ContactStats struct {
	ByStatus []KV `json:"by_status"`
}

type KV struct {
	Key   string `json:"key"`
	Value int    `json:"value"`
}

func (r *ReportsRepo) ContactStats(ctx context.Context) (*ContactStats, error) {
	s := &ContactStats{}
	s.ByStatus, _ = r.queryKV(ctx, `SELECT status, COUNT(*)::int FROM contacts GROUP BY status ORDER BY COUNT(*) DESC`)
	return s, nil
}

type ConversationStats struct {
	ByStatus []KV `json:"by_status"`
	ByInbox  []KV `json:"by_inbox"`
}

func (r *ReportsRepo) ConversationStats(ctx context.Context) (*ConversationStats, error) {
	s := &ConversationStats{}
	s.ByStatus, _ = r.queryKV(ctx, `SELECT status, COUNT(*)::int FROM conversations GROUP BY status`)
	s.ByInbox, _ = r.queryKV(ctx, `SELECT i.name, COUNT(*)::int FROM conversations cv JOIN inboxes i ON i.id=cv.inbox_id GROUP BY i.name`)
	return s, nil
}

func (r *ReportsRepo) AgentStats(ctx context.Context) ([]map[string]any, error) {
	rows, err := r.db.Query(ctx, `
		SELECT u.name,
		       COUNT(cv.id)::int AS total_conversations,
		       COUNT(cv.id) FILTER (WHERE cv.status='resolved')::int AS resolved,
		       COUNT(cv.id) FILTER (WHERE cv.status='open')::int AS open
		FROM users u
		LEFT JOIN conversations cv ON cv.assigned_to=u.id
		GROUP BY u.id, u.name
		ORDER BY total_conversations DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []map[string]any
	for rows.Next() {
		var name string
		var total, resolved, open int
		if err := rows.Scan(&name, &total, &resolved, &open); err != nil {
			return nil, err
		}
		result = append(result, map[string]any{
			"agent": name, "total": total, "resolved": resolved, "open": open,
		})
	}
	return result, nil
}

// ─── Time Series ─────────────────────────────────────────────────────────────

type TimeSeriesPoint struct {
	Timestamp string `json:"timestamp"`
	Value     int    `json:"value"`
}

type TimeSeriesReport struct {
	Conversations    []TimeSeriesPoint `json:"conversations"`
	IncomingMessages []TimeSeriesPoint `json:"incoming_messages"`
	OutgoingMessages []TimeSeriesPoint `json:"outgoing_messages"`
	Resolutions      []TimeSeriesPoint `json:"resolutions"`
}

func (r *ReportsRepo) ConversationTimeSeries(ctx context.Context, from, to time.Time, groupBy string) (*TimeSeriesReport, error) {
	allowed := map[string]string{"day": "day", "week": "week", "month": "month"}
	trunc, ok := allowed[groupBy]
	if !ok {
		trunc = "day"
	}

	dateFormat := "2006-01-02"
	if trunc == "month" {
		dateFormat = "2006-01"
	}

	q := fmt.Sprintf(`
		WITH series AS (
			SELECT generate_series($1::date, $2::date, '1 %s'::interval)::date AS bucket
		)
		SELECT
			s.bucket,
			COUNT(cv.id) FILTER (WHERE cv.created_at::date >= s.bucket AND cv.created_at::date < s.bucket + '1 %s'::interval)::int AS conversations,
			COUNT(m.id) FILTER (WHERE m.direction='incoming' AND m.created_at::date >= s.bucket AND m.created_at::date < s.bucket + '1 %s'::interval)::int AS incoming,
			COUNT(m.id) FILTER (WHERE m.direction='outgoing' AND m.created_at::date >= s.bucket AND m.created_at::date < s.bucket + '1 %s'::interval)::int AS outgoing,
			COUNT(cv2.id) FILTER (WHERE cv2.status='resolved' AND cv2.updated_at::date >= s.bucket AND cv2.updated_at::date < s.bucket + '1 %s'::interval)::int AS resolutions
		FROM series s
		LEFT JOIN conversations cv ON cv.created_at::date >= $1 AND cv.created_at::date <= $2
		LEFT JOIN messages m ON m.created_at::date >= $1 AND m.created_at::date <= $2
		LEFT JOIN conversations cv2 ON cv2.created_at::date >= $1 AND cv2.created_at::date <= $2
		GROUP BY s.bucket
		ORDER BY s.bucket
	`, trunc, trunc, trunc, trunc, trunc)

	rows, err := r.db.Query(ctx, q, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rep := &TimeSeriesReport{
		Conversations:    []TimeSeriesPoint{},
		IncomingMessages: []TimeSeriesPoint{},
		OutgoingMessages: []TimeSeriesPoint{},
		Resolutions:      []TimeSeriesPoint{},
	}

	for rows.Next() {
		var bucket time.Time
		var convs, incoming, outgoing, resolutions int
		if err := rows.Scan(&bucket, &convs, &incoming, &outgoing, &resolutions); err != nil {
			return nil, err
		}
		ts := bucket.Format(dateFormat)
		rep.Conversations = append(rep.Conversations, TimeSeriesPoint{Timestamp: ts, Value: convs})
		rep.IncomingMessages = append(rep.IncomingMessages, TimeSeriesPoint{Timestamp: ts, Value: incoming})
		rep.OutgoingMessages = append(rep.OutgoingMessages, TimeSeriesPoint{Timestamp: ts, Value: outgoing})
		rep.Resolutions = append(rep.Resolutions, TimeSeriesPoint{Timestamp: ts, Value: resolutions})
	}
	return rep, nil
}

func (r *ReportsRepo) queryKV(ctx context.Context, q string) ([]KV, error) {
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []KV
	for rows.Next() {
		var k string
		var v int
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		result = append(result, KV{Key: k, Value: v})
	}
	return result, nil
}
