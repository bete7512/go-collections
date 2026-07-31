package main

func main() {}

type User struct {
	ID   int
	Name string
}

func DedupByID(users []User) []User {

	seen := make(map[int]struct{}, len(users))
	out := make([]User, 0, len(users))

	for _, user := range users {
		if _, ok := seen[user.ID]; ok {
			continue
		}

		seen[user.ID] = struct{}{}

		out = append(out, user)
	}
	return out
}
