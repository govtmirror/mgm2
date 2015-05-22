package core

type subscription interface {
	GetReceive() <-chan UserObject
	Unsubscribe()
}

type sub struct {
	pipe chan UserObject
	del  chan sub
}

func (s sub) Unsubscribe() {
	s.del <- s
}

func (s sub) GetReceive() <-chan UserObject {
	return s.pipe
}

type subscriptionManager interface {
	Subscribe() subscription
	Broadcast(UserObject)
}

type subMgr struct {
	newSubscriptions chan sub
	unsubscribe      chan sub
	broadcast        chan UserObject
}

func newSubscriptionManager() subscriptionManager {
	sm := subMgr{}
	sm.newSubscriptions = make(chan sub, 8)
	sm.unsubscribe = make(chan sub, 8)
	sm.broadcast = make(chan UserObject, 32)
	go sm.process()
	return sm
}

func (sm subMgr) Subscribe() subscription {
	ch := make(chan UserObject, 8)
	s := sub{ch, sm.unsubscribe}
	sm.newSubscriptions <- s
	return s
}

func (sm subMgr) Broadcast(msg UserObject) {
	sm.broadcast <- msg
}

func (sm subMgr) process() {
	var subscribers []sub
	for {
		select {
		case s := <-sm.newSubscriptions:
			subscribers = append(subscribers, s)
		case s := <-sm.unsubscribe:
			for i, sub := range subscribers {
				if sub == s {
					subscribers[i] = subscribers[len(subscribers)-1]
					subscribers = subscribers[:len(subscribers)-1]
				}
			}
		case ob := <-sm.broadcast:
			for _, s := range subscribers {
				go func(s sub, ob UserObject) {
					s.pipe <- ob
				}(s, ob)
			}
		}
	}
}
