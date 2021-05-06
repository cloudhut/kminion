package e2e

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

type EndToEndMessage struct {
	MinionID  string  `json:"minionID"`
	Timestamp float64 `json:"timestamp"`
}

func (s *Service) produceToManagementTopic(ctx context.Context) error {

	topicName := s.config.TopicManagement.Name

	record, err := createEndToEndRecord(topicName, s.minionID)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			startTime := timeNowMs()
			s.endToEndMessagesProduced.Inc()

			err = s.client.Produce(ctx, record, func(r *kgo.Record, err error) {
				endTime := timeNowMs()
				ackDurationMs := endTime - startTime
				ackDuration := time.Duration(ackDurationMs) * time.Millisecond

				if err != nil {
					s.logger.Error("error producing record: %w", zap.Error(err))
				} else {
					s.onAck(r.Partition, ackDuration)
				}
			})

			if err != nil {
				return err
			}
			return nil
		}
	}

}

func createEndToEndRecord(topicName string, minionID string) (*kgo.Record, error) {

	timestamp := timeNowMs()
	message := EndToEndMessage{
		MinionID:  minionID,
		Timestamp: timestamp,
	}
	mjson, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	record := &kgo.Record{
		Topic: topicName,
		Key:   []byte(uuid.NewString()),
		Value: []byte(mjson),
	}

	return record, nil
}