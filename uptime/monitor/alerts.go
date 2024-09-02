package monitor

import (
	"context"
	"encore.app/site"
	"encore.dev/pubsub"
	"encore.dev/storage/sqldb"
	"errors"
)

type TransitionEvent struct {
	Site *site.Site `json:"site"`
	Up   bool       `json:"up"`
}

var TransitionTopic = pubsub.NewTopic[*TransitionEvent]("uptime-transition", pubsub.TopicConfig{DeliveryGuarantee: pubsub.AtLeastOnce})

func getPreviousMeasurement(ctx context.Context, siteID int) (up bool, err error) {
	err = db.QueryRow(ctx, `
		SELECT up FROM checks
		WHERE site_id = $1
		ORDER BY checked_at DESC
		LIMIT 1
	`, siteID).Scan(&up)

	if errors.Is(err, sqldb.ErrNoRows) {
		// There was no previous ping; treat this as if the site was up before
		return true, nil
	} else if err != nil {
		return false, err
	}
	return up, nil
}

func publishOnTransition(ctx context.Context, site *site.Site, isUp bool) error {
	wasUp, err := getPreviousMeasurement(ctx, site.ID)
	if err != nil {
		return err
	}
	if isUp == wasUp {
		// Nothing to do
		return nil
	}
	_, err = TransitionTopic.Publish(ctx, &TransitionEvent{
		Site: site,
		Up:   isUp,
	})
	return err
}
