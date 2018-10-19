package context

type MockContext struct {
}

func (m *MockContext) Publish(*R) {

}
func (m *MockContext) Request() *R {
	return &R{Headers: M{}}
}
