package yak

import (
	"github.com/davecgh/go-spew/spew"
	"github.com/yaklang/yaklang/common/fp"
	"github.com/yaklang/yaklang/common/yakgrpc/ypb"
	"testing"
)

func TestMixPluginCaller_SetFeedback(t *testing.T) {
	manager, err := NewMixPluginCaller()
	_ = err
	manager.SetConcurrent(20)
	manager.SetDividedContext(true)

	manager.SetFeedback(func(i *ypb.ExecResult) error {
		spew.Dump(i.Message)
		return nil
	})

	err = manager.LoadPlugin("Set Feedback")
	if err != nil {
		t.FailNow()
		return
	}

	manager.HandleServiceScanResult(&fp.MatchResult{
		Target:      "127.0.0.1",
		Port:        1111,
		State:       "OPEN",
		Reason:      "xxxxxxxxxxxxxxx",
		Fingerprint: nil,
	})
	manager.Wait()

}
