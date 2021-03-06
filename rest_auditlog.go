package disgord

import (
	"errors"
	"strconv"

	"github.com/andersfylling/disgord/endpoint"
	"github.com/andersfylling/disgord/httd"
)

// GuildAuditLogsParams set params used in endpoint request
// https://discordapp.com/developers/docs/resources/audit-log#get-guild-audit-log-query-string-parameters
type GuildAuditLogsParams struct {
	UserID     Snowflake `urlparam:"user_id,omitempty"`     // filter the log for a user id
	ActionType uint      `urlparam:"action_type,omitempty"` // the type of audit log event
	Before     Snowflake `urlparam:"before,omitempty"`      // filter the log before a certain entry id
	Limit      int       `urlparam:"limit,omitempty"`       // how many entries are returned (default 50, minimum 1, maximum 100)
}

// GetQueryString .
func (params *GuildAuditLogsParams) GetQueryString() string {
	separator := "?"
	query := ""

	if !params.UserID.Empty() {
		query += separator + "user_id=" + params.UserID.String()
		separator = "&"
	}

	if params.ActionType > 0 {
		query += separator + "action_type=" + strconv.FormatUint(uint64(params.ActionType), 10)
		separator = "&"
	}

	if !params.Before.Empty() {
		query += separator + "before=" + params.Before.String()
		separator = "&"
	}

	if params.Limit > 0 {
		query += separator + "limit=" + strconv.Itoa(params.Limit)
	}

	return query
}

// GuildAuditLogs [REST] Returns an audit log object for the guild. Requires the 'VIEW_AUDIT_LOG' permission.
//  Method                   GET
//  Endpoint                 /guilds/{guild.id}/audit-logs
//  Rate limiter [MAJOR]     /guilds/{guild.id}/audit-logs
//  Discord documentation    https://discordapp.com/developers/docs/resources/audit-log#get-guild-audit-log
//  Reviewed                 2018-06-05
//  Comment                  -
//  Note                     Check the last entry in the cache, to avoid fetching data we already got
func GuildAuditLogs(client httd.Getter, guildID Snowflake, params *GuildAuditLogsParams) (log *AuditLog, err error) {
	details := &httd.Request{
		Ratelimiter: ratelimitGuildAuditLogs(guildID),
		Endpoint:    endpoint.GuildAuditLogs(guildID) + params.GetQueryString(),
	}
	resp, body, err := client.Get(details)
	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = errors.New("incorrect status code. Got " + strconv.Itoa(resp.StatusCode) + ", wants 200. Message: " + string(body))
		return
	}

	err = unmarshal(body, &log)
	return
}
