package tgSend

import "testing"

func TestSend(t *testing.T) {
	type args struct {
		proxyAddress string
		chatID       int64
		text         string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{"test1", args{"http://127.0.0.1:7890", 956772010, "这是一条消息"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Send(tt.args.proxyAddress, tt.args.chatID, tt.args.text); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
