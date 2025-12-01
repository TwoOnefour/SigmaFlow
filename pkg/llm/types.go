package llm

type Messages struct {
	Role    Role
	Content string
}

type Role string

const (
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
)

func (r Role) ToString() string {
	return string(r)
}
