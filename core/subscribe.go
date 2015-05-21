package core

type subscription struct {
	pipe chan UserObject
	del  chan chan UserObject
}

func (s subscription) unsubscribe() {
	s.del <- s.pipe
}

type subscriptionManager struct {
	newSubscriptions chan chan UserObject
	unsubscribe      chan chan UserObject
	broadcast        chan UserObject
}

func newSubscriptionManager() subscriptionManager {
	sm := subscriptionManager{}
	sm.newSubscriptions = make(chan chan UserObject, 8)
	sm.unsubscribe = make(chan chan UserObject, 8)
	go sm.process()
	return sm
}

func (sm subscriptionManager) Subscribe() subscription {
	ch := make(chan UserObject, 8)
	sm.newSubscriptions <- ch
	return subscription{ch, sm.unsubscribe}
}

func (sm subscriptionManager) process() {
	var subscribers []chan UserObject
	for {
		select {
		case s := <-sm.newSubscriptions:
			subscribers = append(subscribers, s)
		case s := <-sm.unsubscribe:
			for i, sub := range subscribers {
				if sub == s {
					subscribers[i] = subscribers[len(subscribers)-1]
					subscribers = subscribers[:len(subscribers)-1]
					break
				}
			}
		case ob := <-sm.broadcast:
			for _, sub := range subscribers {
				go func(ch chan<- UserObject, ob UserObject) {
					ch <- ob
				}(sub, ob)
			}
		}
	}
}
