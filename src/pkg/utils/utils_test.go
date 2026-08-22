package utils

import (
	"testing"
	"time"
)

// formatMs 毫秒时间戳格式化
func formatMs(ms int64) string {
	return time.UnixMilli(ms).Format("2006-01-02 15:04:05")
}

// go test -v ./src/pkg/utils -run TestCalcOneThirdTimePoint

func TestCalcOneThirdTimePoint(t *testing.T) {
	baseTime, _ := time.Parse("2006-01-02 15:04:05", "2026-08-23 10:00:00")
	nowMock := baseTime.UnixMilli()

	tests := []struct {
		name         string
		expTs        int64
		mockNowTs    int64
		wantTargetTs int64
	}{
		{
			name:         "2小时后过期",
			expTs:        baseTime.Add(2 * time.Hour).UnixMilli(),    // 12:00:00
			mockNowTs:    nowMock,                                    // 10:00:00
			wantTargetTs: baseTime.Add(40 * time.Minute).UnixMilli(), // 10:40:00
		},
		{
			name:         "token已经过期",
			expTs:        baseTime.Add(-1 * time.Hour).UnixMilli(),
			mockNowTs:    nowMock,
			wantTargetTs: nowMock,
		},
		{
			name:         "30分钟后过期",
			expTs:        baseTime.Add(30 * time.Minute).UnixMilli(), //10:30
			mockNowTs:    nowMock,                                    //10:00
			wantTargetTs: baseTime.Add(10 * time.Minute).UnixMilli(), //10:10
		},
		{
			name:         "剩余时间不能被3整除，整数向下取整",
			expTs:        baseTime.Add(31 * time.Minute).UnixMilli(),
			mockNowTs:    nowMock,
			wantTargetTs: baseTime.Add(10*time.Minute + 20*time.Second).UnixMilli(),
		},
		{
			name:         "刚好等于当前时间，已过期",
			expTs:        nowMock,
			mockNowTs:    nowMock,
			wantTargetTs: nowMock,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calcOneThirdTimePoint(tt.expTs, tt.mockNowTs)

			t.Logf("now      : %s", formatMs(tt.mockNowTs))
			t.Logf("expire   : %s", formatMs(tt.expTs))
			t.Logf("got point: %s", formatMs(got))
			t.Logf("want point: %s", formatMs(tt.wantTargetTs))

			if got != tt.wantTargetTs {
				t.Errorf("calcOneThirdTimePoint() got = %d, want = %d", got, tt.wantTargetTs)
			}
		})
	}
}
