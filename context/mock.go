package context

type MockContext struct {
	Response *R
}

func (m *MockContext) Ack(...interface{}) {

}

func (m *MockContext) Nack(...interface{}) {

}

func (m *MockContext) Publish(r *R) {
	m.Response = r
}
func (m *MockContext) Request() *R {
	return &R{Headers: M{}}
}
