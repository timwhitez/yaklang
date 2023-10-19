package match

import (
	"github.com/google/gopacket"
	"github.com/yaklang/yaklang/common/suricata/rule"
)

type GroupOption func(group *Group)

func WithGroupOnMatchedCallback(cb func(packet gopacket.Packet, match *rule.Rule)) GroupOption {
	return func(c *Group) {
		c.onMatchedCallback = cb
	}
}

func WithOnRuleLoad(cb func(match *rule.Rule) error) GroupOption {
	return func(c *Group) {
		c.onRuleLoadCallback = cb
	}
}
