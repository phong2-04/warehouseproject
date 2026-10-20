
package model

type User struct {
    ID           int64  `json:"id"`
    Username     string `json:"username"`
    PasswordHash string `json:"-"`
    FullName     string `json:"full_name"`
    Role         string `json:"role"`
}
