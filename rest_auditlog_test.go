package disgord

import (
	"testing"
)

func TestAuditLogParams(t *testing.T) {
	params := &GuildAuditLogsParams{}
	var wants string

	wants = ""
	verifyQueryString(t, params, wants)

	s := "438543957"
	params.UserID, _ = GetSnowflake(s)
	wants = "?user_id=" + s
	verifyQueryString(t, params, wants)

	params.ActionType = 6
	wants += "&action_type=6"
	verifyQueryString(t, params, wants)

	params.ActionType = 0
	wants = "?user_id=" + s
	verifyQueryString(t, params, wants)
}

func TestGuildAuditLogs(t *testing.T) {
	client, keys, err := createTestRequester()
	if err != nil {
		t.Skip()
		return
	}

	params := &GuildAuditLogsParams{}
	log, err := GuildAuditLogs(client, keys.GuildAdmin, params)
	if err != nil {
		t.Error(err)
	}

	if log == nil {
		t.Error("did not get a datastructure from rest.GuildAuditLogs()")
	}
}
