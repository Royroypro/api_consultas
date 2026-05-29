package main
import "golang.org/x/crypto/bcrypt"
import "fmt"
func main() {
    hash, _ := bcrypt.GenerateFromPassword([]byte("admin2026"), bcrypt.DefaultCost)
    fmt.Println(string(hash))
    err := bcrypt.CompareHashAndPassword([]byte("$2a$10$wE1m.hVdSw82iFqf9zP65e4xIlgfG/Yy7L1gZ2o2Nq2.u/b7jYee."), []byte("admin2026"))
    fmt.Println(err)
}
