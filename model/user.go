package model

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/scrypt"
)

type User struct {
	gorm.Model
	Username       string `gorm:"unique"`
	HashedPassword string
	Salt           string
	Email          *string `gorm:"unique"`
	EmailVerified  bool
	FacebookID     *string `gorm:"unique"`
	Role           string
	//TODO ExpiredPassword bool
}

func (m *Model) RetrieveUserFB(facebookid string) *User {
	var u User
	res := m.Db.First(&u, "facebook_id = ?", facebookid)
	if res.Error != nil {
		return nil
	}
	return &u
}

func (m *Model) LoginOrCreateFB(currentUser *User, facebookID string, name string, maybeEmail *string) (*User, error) {
	// if that facebookID is already in the system, we want to log the user in with that
	existingFacebookUser := m.RetrieveUserFB(facebookID)
	if existingFacebookUser != nil {
		// Update user email if needed. Mark it as verified
		if maybeEmail != nil {
			if existingFacebookUser.Email != maybeEmail {
				existingFacebookUser.Email = maybeEmail
			}
			existingFacebookUser.EmailVerified = true
			err := m.Db.Save(&existingFacebookUser).Error
			if err != nil {
				return nil, err
			}
		}
		return existingFacebookUser, nil
	}

	// If the facebookID is new for us, but we are already logged in
	// then we want to add the facebook data to the existing profile
	if currentUser != nil {
		currentUser.FacebookID = &facebookID
		err := m.Db.Save(&currentUser).Error
		if err != nil {
			return nil, err
		}
		return currentUser, nil
	}

	// If we are not logged in, and the facebookID is new, then we want to
	// check if the user is already registered by his facebook mail
	if maybeEmail != nil {
		var existingEmailUser User
		err := m.Db.First(&existingEmailUser, "email = ? and email_verified = 1", maybeEmail).Error
		if err == nil {
			existingEmailUser.FacebookID = &facebookID
			existingEmailUser.EmailVerified = true

			err = m.Db.Save(&existingEmailUser).Error
			if err != nil {
				return nil, err
			}
		}
	}

	// We are not logged in, and both the facebookID and the mail are not
	// in our system, we create a new User
	newUser := &User{
		Username:   name,
		FacebookID: &facebookID,
		Role:       "user",
	}

	if maybeEmail != nil {
		newUser.Email = maybeEmail
		newUser.EmailVerified = true
	}

	err := m.Db.Save(newUser).Error
	if err != nil {
		return nil, err
	}
	return newUser, nil
}

func (m *Model) UserPage(u *User) *Page {
	var page Page
	res := m.Db.First(&page, "owner_id = ? and type = ?", u.ID, PageUser)
	if res.Error != nil {
		return nil
	}
	return &page
}

func (m *Model) RetrieveUser(id uint) *User {
	var u User
	res := m.Db.First(&u, "id = ?", id)
	if res.Error != nil {
		return nil
	}
	return &u
}

func (m *Model) LoginEmail(email string, password string) *User {
	var u User
	res := m.Db.First(&u, "email = ?", email)
	if res.Error != nil {
		return nil
	}

	if u.CheckPassword(password) == false {
		return nil
	}

	return &u
}

func (m *Model) RegisterEmail(username string, email string, password string, role string) (*User, error) {
	u := User{
		Username:      username,
		Email:         &email,
		Role:          role,
		EmailVerified: true,
	}
	u.SetPassword(password)
	res := m.Db.Save(&u)
	if res.Error != nil {
		return nil, res.Error
	}
	return &u, nil
}

func (u *User) SetPassword(password string) {
	salt := make([]byte, 8)
	_, err := rand.Read(salt)
	if err != nil {
		panic(err)
	}
	u.Salt = string(salt)
	u.HashedPassword = HashPassword(password, salt)
}

func (u *User) CheckPassword(password string) bool {
	if password == "" {
		return false
	}
	return HashPassword(password, []byte(u.Salt)) == u.HashedPassword
}

func HashPassword(password string, salt []byte) string {
	dk, err := scrypt.Key([]byte(password), salt, 32768, 8, 1, 32)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(dk)
}

/* For future use
func (m *Model) MakeAdminByMail(email string, role string) (int64, error) {
	res := m.Db.Exec("UPDATE users SET role=? WHERE email=?", role, email)
	return res.RowsAffected, res.Error
} */
