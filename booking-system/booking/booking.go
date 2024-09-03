package booking

import (
	"context"
	"encore.app/booking/db"
	"encore.app/sendgrid"
	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"slices"
	"time"
)

var (
	bookingDB = sqldb.NewDatabase("booking", sqldb.DatabaseConfig{
		Migrations: "./db/migrations",
	})

	pgxdb = sqldb.Driver[*pgxpool.Pool](bookingDB)
	query = db.New(pgxdb)
)

type Booking struct {
	ID    int64     `json:"id"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Email string    `encore:"sensitive"`
}

type BookParams struct {
	Start time.Time `json:"start"`
	Email string    `encore:"sensitive"`
}

//encore:api public method=POST path=/booking
func Book(ctx context.Context, p *BookParams) error {
	eb := errs.B()

	now := time.Now()
	if p.Start.Before(now) {
		return eb.Code(errs.InvalidArgument).Msg("start time must be in the future").Err()
	}

	tx, err := pgxdb.Begin(ctx)
	if err != nil {
		return eb.Cause(err).Code(errs.Unavailable).Msg("failed to start transaction").Err()
	}
	defer tx.Rollback(context.Background()) // committed explicitly on success

	// Get the bookings for this day.
	startOfDay := time.Date(p.Start.Year(), p.Start.Month(), p.Start.Day(), 0, 0, 0, 0, p.Start.Location())
	bookings, err := listBookingsBetween(ctx, startOfDay, startOfDay.AddDate(0, 0, 1))
	if err != nil {
		return eb.Cause(err).Code(errs.Unavailable).Msg("failed to list bookings").Err()
	}

	// Is this slot bookable?
	slot := BookableSlot{Start: p.Start, End: p.Start.Add(DefaultBookingDuration)}
	if len(filterBookableSlots([]BookableSlot{slot}, now, bookings)) == 0 {
		return eb.Code(errs.InvalidArgument).Msg("slot is unavailable").Err()
	}

	_, err = query.InsertBooking(ctx, db.InsertBookingParams{
		StartTime: pgtype.Timestamp{Time: p.Start, Valid: true},
		EndTime:   pgtype.Timestamp{Time: p.Start.Add(DefaultBookingDuration), Valid: true},
		Email:     p.Email,
	})
	if err != nil {
		return eb.Cause(err).Code(errs.Unavailable).Msg("failed to insert booking").Err()
	}

	if err := tx.Commit(ctx); err != nil {
		return eb.Cause(err).Code(errs.Unavailable).Msg("failed to commit transaction").Err()
	}

	// Send confirmation email using Sendgrid
	formattedTime := pgtype.Timestamp{Time: p.Start, Valid: true}.Time.Format("2006-01-02 15:04")
	_, err = sendgrid.Send(ctx, &sendgrid.SendParams{
		From: sendgrid.Address{
			Name:  "<your name>",
			Email: "<your email>",
		},
		To: sendgrid.Address{
			Email: p.Email,
		},
		Subject: "Booking Confirmation",
		Text:    "Thank you for your booking!\nWe look forward to seeing you soon at " + formattedTime,
		Html:    "",
	})

	if err != nil {
		return err
	}

	return nil
}

func listBookingsBetween(
	ctx context.Context,
	start, end time.Time,
) ([]*Booking, error) {
	rows, err := query.ListBookingsBetween(ctx, db.ListBookingsBetweenParams{
		StartTime: pgtype.Timestamp{Time: start, Valid: true},
		EndTime:   pgtype.Timestamp{Time: end, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	var bookings []*Booking
	for _, row := range rows {
		bookings = append(bookings, &Booking{
			ID:    row.ID,
			Start: row.StartTime.Time,
			End:   row.EndTime.Time,
			Email: row.Email,
		})
	}
	return bookings, nil
}

func filterBookableSlots(
	slots []BookableSlot,
	now time.Time,
	bookings []*Booking,
) []BookableSlot {
	// Remove slots for which the start time has already passed.
	slots = slices.DeleteFunc(slots, func(s BookableSlot) bool {
		// Has the slot already passed?
		if s.Start.Before(now) {
			return true
		}

		// Is there a booking that overlaps with this slot?
		for _, b := range bookings {
			if b.Start.Before(s.End) && b.End.After(s.Start) {
				return true
			}
		}

		return false
	})
	return slots
}
