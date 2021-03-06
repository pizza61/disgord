package disgord

import (
	"github.com/andersfylling/disgord/cache/interfaces"
)

func createChannelCacher(conf *CacheConfig) (cacher interfaces.CacheAlger, err error) {
	if conf.DisableChannelCaching {
		return nil, nil
	}

	const channelWeight = 1 // MiB. TODO: what is the actual max size?
	limit := conf.ChannelCacheLimitMiB / channelWeight

	cacher, err = constructSpecificCacher(conf.ChannelCacheAlgorithm, limit, conf.ChannelCacheLifetime)
	return
}

type channelCacheItem struct {
	channel *Channel
}

func (c *channelCacheItem) process(channel *Channel, immutable bool) {
	if immutable {
		c.channel = channel.DeepCopy().(*Channel)
		c.channel.Recipients = []*User{} // clear
	} else {
		c.channel = channel
	}

	c.channel.recipientsIDs = make([]Snowflake, len(channel.Recipients))
	for i := range channel.Recipients {
		c.channel.recipientsIDs = append(c.channel.recipientsIDs, channel.Recipients[i].ID)
	}
}

func (c *channelCacheItem) build(cache *Cache) (channel *Channel) {
	if cache.immutable {
		channel = c.channel.DeepCopy().(*Channel)
	} else {
		channel = c.channel
	}

	if channel.Type != ChannelTypeDM && channel.Type != ChannelTypeGroupDM {
		return
	}

	recipients := make([]*User, len(channel.recipientsIDs))
	for i := range c.channel.recipientsIDs {
		usr, err := cache.GetUser(c.channel.recipientsIDs[i]) // handles immutability on it's own
		if err != nil || usr == nil {
			usr = NewUser()
			usr.ID = c.channel.recipientsIDs[i]
			// TODO: should this be loaded by REST request?...
			// TODO-2: maybe it can be a cache option to load dead members on read?
		}
		recipients[i] = usr
	}

	// TODO-racecondition: when !immutable
	channel.Recipients = recipients
	return
}

func (c *channelCacheItem) update(fresh *Channel, immutable bool) {
	if !immutable {
		c.channel = fresh
		return
	}

	fresh.copyOverToCache(c.channel)
}

// SetChannel adds a new channel to cache or updates an existing one
func (c *Cache) SetChannel(new *Channel) {
	if c.channels == nil || new == nil {
		return
	}

	c.channels.Lock()
	defer c.channels.Unlock()
	if item, exists := c.channels.Get(new.ID); exists {
		item.Object().(*channelCacheItem).update(new, c.immutable)
		c.channels.RefreshAfterDiscordUpdate(item)
	} else {
		content := &channelCacheItem{}
		content.process(new, c.immutable)
		c.channels.Set(new.ID, c.channels.CreateCacheableItem(content))
	}
}

// UpdateChannelPin ...
func (c *Cache) UpdateChannelPin(id Snowflake, timestamp Timestamp) {
	if c.channels == nil || id.Empty() {
		return
	}

	c.channels.Lock()
	defer c.channels.Unlock()
	if item, exists := c.channels.Get(id); exists {
		item.Object().(*channelCacheItem).channel.LastPinTimestamp = timestamp
		c.channels.RefreshAfterDiscordUpdate(item)
	} else {
		// channel does not exist in cache, create a partial channel
		partial := &Channel{ID: id, LastPinTimestamp: timestamp}
		content := &channelCacheItem{}
		content.process(partial, c.immutable)
		c.channels.Set(id, c.channels.CreateCacheableItem(content))
	}
}

// UpdateChannelLastMessageID ...
func (c *Cache) UpdateChannelLastMessageID(channelID Snowflake, messageID Snowflake) {
	if c.channels == nil || channelID.Empty() || messageID.Empty() {
		return
	}

	c.channels.Lock()
	defer c.channels.Unlock()
	if item, exists := c.channels.Get(channelID); exists {
		item.Object().(*channelCacheItem).channel.LastMessageID = messageID
		c.channels.RefreshAfterDiscordUpdate(item)
	} else {
		// channel does not exist in cache, create a partial channel
		// this is an indirect channel update..
		//partial := &PartialChannel{ID: channelID, LastMessageID: messageID}
		//content := &channelCacheItem{}
		//content.process(partial, c.immutable)
		//c.channels.Set(channelID, c.channels.CreateCacheableItem(content))
	}
}

// GetChannel ...
func (c *Cache) GetChannel(id Snowflake) (channel *Channel, err error) {
	if c.channels == nil {
		err = newErrorUsingDeactivatedCache("channels")
		return
	}

	c.channels.RLock()
	defer c.channels.RUnlock()

	var exists bool
	var result interfaces.CacheableItem
	if result, exists = c.channels.Get(id); !exists {
		err = newErrorCacheItemNotFound(id)
		return
	}

	channel = result.Object().(*channelCacheItem).build(c)
	return
}

// DeleteChannel ...
func (c *Cache) DeleteChannel(id Snowflake) {
	c.channels.Lock()
	defer c.channels.Unlock()

	c.channels.Delete(id)
}
