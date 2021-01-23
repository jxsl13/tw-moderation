package mqtt

import (
	"testing"
	"time"
)

func TestNewPublisherSubscriber(t *testing.T) {
	type args struct {
		address        string
		clientID       string
		publishTopic   string
		subscribeTopic string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"Test PubSub #1",
			args{
				"tcp://localhost:1883",
				"PubSubTest#1",
				"PubSubTest#1#Topic",
				"PubSubTest#1#Topic",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pub, sub, err := NewPublisherSubscriber(tt.args.address, tt.args.clientID, tt.args.publishTopic, tt.args.subscribeTopic)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewPublisherSubscriber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			str := "test"
			go func() {
				time.Sleep(time.Second * 5)
				pub.Publish(str)
			}()

			t.Logf("Published %s", str)
			if result := <-sub.Next(); result != str {
				t.Errorf("wanted: %s got: %s", str, result)
			} else {
				t.Logf("Received: ")
			}

			pub.Close()
			sub.Close()
		})
	}
}
