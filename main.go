package main

import (
	"database/sql"
	"html/template"
	"net/http"

	"code.google.com/p/go.crypto/bcrypt"
	"github.com/go-martini/martini"
	_ "github.com/lib/pq"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/sessions"
	"github.com/russross/blackfriday"
)

type Article struct {
	Id      int
	Title   string
	Content string
}

type User struct {
	Name  string
	Email string
}

func SetupDB() *sql.DB {
	db, err := sql.Open("postgres", "dbname=blog sslmode=disable")
	PanicIf(err)

	return db
}

func PanicIf(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	m := martini.Classic()
	m.Map(SetupDB())
	m.Use(render.Renderer(render.Options{
		Layout: "layout",
		Funcs: []template.FuncMap{
			{
				"unescaped": func(args ...interface{}) template.HTML {
					return template.HTML(args[0].(string))
				},
			},
		},
	}))

	//Sessions
	store := sessions.NewCookieStore([]byte("secret123"))
	m.Use(sessions.Sessions("vsauth", store))

	m.Get("/", IndexArticles)
	m.Get("/create", CreateArticle)
	m.Get("/read", ReadArticle)
	m.Get("/update", UpdateArticle)
	m.Get("/delete", DeleteArticle)
	m.Post("/save", SaveArticle)
	m.Get("/login", LoginForm)
	m.Post("/login", UserLogin)
	m.Get("/signup", SignupForm)
	m.Post("/signup", UserSignup)
	m.Get("/logout", UserLogout)
	m.Get("/wrong", WrongLogin)

	m.Run()
}

func IndexArticles(db *sql.DB, r *http.Request, ren render.Render) {
	search := "%" + r.URL.Query().Get("search") + "%"
	rows, err := db.Query(`SELECT id, title, content
                         FROM article
                         WHERE title ILIKE $1
                         OR content ILIKE $1
												 ORDER BY id DESC`, search)
	PanicIf(err)
	defer rows.Close()

	articles := []Article{}
	for rows.Next() {
		a := Article{}
		err := rows.Scan(&a.Id, &a.Title, &a.Content)
		a.Content = PrepareContent(a.Content)

		PanicIf(err)
		articles = append(articles, a)
	}

	if len(articles) == 0 {
		a := Article{}
		a.Content = "No posts found"
		articles = append(articles, a)
	}

	ren.HTML(200, "articles", articles)
}

func CreateArticle(ren render.Render, s sessions.Session, db *sql.DB, c martini.Context) {
	RequireLogin(s, db, c, ren, "/create")
	ren.HTML(200, "form", nil)
}

func ReadArticle(db *sql.DB, r *http.Request, ren render.Render) {
	id := r.URL.Query().Get("id")
	a := &Article{}
	err := db.QueryRow(`SELECT id, title, content
												FROM article
												WHERE id = $1`, id).Scan(&a.Id, &a.Title, &a.Content)
	a.Content = PrepareContent(a.Content)
	PanicIf(err)

	ren.HTML(200, "article", a)
}

func UpdateArticle(db *sql.DB, r *http.Request, ren render.Render, s sessions.Session, c martini.Context) {
	id := r.URL.Query().Get("id")
	RequireLogin(s, db, c, ren, "/update?id="+id)

	a := &Article{}
	err := db.QueryRow(`SELECT id, title, content
												FROM article
												WHERE id = $1`, id).Scan(&a.Id, &a.Title, &a.Content)
	PanicIf(err)

	ren.HTML(200, "form", a)
}

func DeleteArticle(db *sql.DB, r *http.Request, rw http.ResponseWriter) {
	id := r.URL.Query().Get("id")
	_, err := db.Exec("DELETE FROM article WHERE id = $1", id)
	PanicIf(err)

	//redirect to home page
	http.Redirect(rw, r, "/", http.StatusFound)
}

func SaveArticle(ren render.Render, r *http.Request, db *sql.DB) {

	id := r.FormValue("articleId")

	if id == "" {
		rows, err := db.Query("INSERT INTO article (title, content) VALUES ($1, $2)",
			r.FormValue("title"),
			r.FormValue("content"))

		PanicIf(err)
		defer rows.Close()

	} else {
		rows, err := db.Query("UPDATE article SET title = $1, content = $2 WHERE id = $3",
			r.FormValue("title"),
			r.FormValue("content"),
			id)

		PanicIf(err)
		defer rows.Close()
	}

	ren.Redirect("/")
}

func LoginForm(ren render.Render) {
	ren.HTML(200, "loginform", nil)
}

func UserLogin(r *http.Request, db *sql.DB, s sessions.Session, rw http.ResponseWriter) (int, string) {
	var id string
	var pass string

	email, password := r.FormValue("email"), r.FormValue("password")
	err := db.QueryRow("select id, password from appuser where email=$1", email).Scan(&id, &pass)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(pass), []byte(password)) != nil {
		//return 401, "Not Authorized. Buuuurn!"
		http.Redirect(rw, r, "/wrong", http.StatusFound)
	}

	//set the user id in the session
	s.Set("userId", id)

	//return user
	if returnUrl, ok := s.Get("returnUrl").(string); ok {
		s.Delete("returnUrl")
		http.Redirect(rw, r, returnUrl, http.StatusFound)
	} else {
		http.Redirect(rw, r, "/", http.StatusFound)
	}

	return 200, "User id is " + id
}

func WrongLogin(ren render.Render) {
	ren.HTML(401, "wronglogin", nil)
}

func SignupForm(ren render.Render) {
	ren.HTML(200, "signupform", nil)
}

func UserSignup(rw http.ResponseWriter, r *http.Request, db *sql.DB) {
	name, email, password := r.FormValue("name"), r.FormValue("email"), r.FormValue("password")

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	PanicIf(err)

	_, err = db.Exec("insert into appuser (name, email, password) values ($1, $2, $3)", name, email, hashedPassword)
	PanicIf(err)

	//redirect to login screen
	http.Redirect(rw, r, "/login", http.StatusFound)
}

func UserLogout(ren render.Render, s sessions.Session) {
	s.Delete("userId")
	ren.HTML(200, "logout", nil)
}

func RequireLogin(s sessions.Session, db *sql.DB, c martini.Context, ren render.Render, returnUrl string) {
	user := &User{}
	err := db.QueryRow("select name, email from appuser where id=$1", s.Get("userId")).Scan(&user.Name, &user.Email)

	s.Set("returnUrl", returnUrl)

	if err != nil {
		ren.Redirect("/login")
		return
	}

	//map user to the context
	c.Map(user)
}

func PrepareContent(content string) string {
	body := string(blackfriday.MarkdownBasic([]byte(content)))
	return body
}

func Label() string {
	return "This is call! "
}
