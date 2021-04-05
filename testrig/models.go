package testrig

import (
	"crypto/rand"
	"crypto/rsa"
	"net"
	"time"

	"github.com/superseriousbusiness/gotosocial/internal/db/model"
	"github.com/superseriousbusiness/gotosocial/internal/oauth"
)

func TestTokens() map[string]*oauth.Token {
	tokens := map[string]*oauth.Token{
		"local_account_1": {

		},
	}
	return tokens
}

func TestClients() map[string]*oauth.Client {
	clients := map[string]*oauth.Client{
		"local_account_1": {
			ID:     "73b48d42-029d-4487-80fc-329a5cf67869",
			Secret: "c3724c74-dc3b-41b2-a108-0ea3d8399830",
			Domain: "http://localhost:8080",
			UserID: "44e36b79-44a4-4bd8-91e9-097f477fe97b", // local_account_1
		},
	}
	return clients
}

func TestApplications() map[string]*model.Application {
	apps := map[string]*model.Application{
		"application_1": {
			ID:           "f88697b8-ee3d-46c2-ac3f-dbb85566c3cc",
			Name:         "really cool gts application",
			Website:      "https://reallycool.app",
			RedirectURI:  "http://localhost:8080",
			ClientID:     "73b48d42-029d-4487-80fc-329a5cf67869", // client_1
			ClientSecret: "c3724c74-dc3b-41b2-a108-0ea3d8399830", // client_1
			Scopes:       "read write follow push",
			VapidKey:     "4738dfd7-ca73-4aa6-9aa9-80e946b7db36",
		},
	}
	return apps
}

func TestUsers() map[string]*model.User {
	users := map[string]*model.User{
		"unconfirmed_account": {
			ID:                     "0f7b1d24-1e49-4ee0-bc7e-fd87b7289eea",
			Email:                  "",
			AccountID:              "59e197f5-87cd-4be8-ac7c-09082ccc4b4d",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              time.Now(),
			SignUpIP:               net.ParseIP("199.222.111.89"),
			UpdatedAt:              time.Time{},
			CurrentSignInAt:        time.Time{},
			CurrentSignInIP:        nil,
			LastSignInAt:           time.Time{},
			LastSignInIP:           nil,
			SignInCount:            0,
			InviteID:               "",
			ChosenLanguages:        []string{},
			FilteredLanguages:      []string{},
			Locale:                 "en",
			CreatedByApplicationID: "",
			LastEmailedAt:          time.Time{},
			ConfirmationToken:      "a5a280bd-34be-44a3-8330-a57eaf61b8dd",
			ConfirmedAt:            time.Time{},
			ConfirmationSentAt:     time.Now(),
			UnconfirmedEmail:       "weed_lord420@example.org",
			Moderator:              false,
			Admin:                  false,
			Disabled:               false,
			Approved:               false,
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
		"admin_account": {
			ID:                     "0fb02eae-2214-473f-9667-0a43f22d75ff",
			Email:                  "admin@example.org",
			AccountID:              "8020dbb4-1e7b-4d99-a872-4cf94e64210f",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              time.Now().Add(-72 * time.Hour),
			SignUpIP:               net.ParseIP("89.22.189.19"),
			UpdatedAt:              time.Now().Add(-72 * time.Hour),
			CurrentSignInAt:        time.Now().Add(-10 * time.Minute),
			CurrentSignInIP:        net.ParseIP("89.122.255.1"),
			LastSignInAt:           time.Now().Add(-2 * time.Hour),
			LastSignInIP:           net.ParseIP("89.122.255.1"),
			SignInCount:            78,
			InviteID:               "",
			ChosenLanguages:        []string{"en"},
			FilteredLanguages:      []string{},
			Locale:                 "en",
			CreatedByApplicationID: "",
			LastEmailedAt:          time.Now().Add(-30 * time.Minute),
			ConfirmationToken:      "",
			ConfirmedAt:            time.Time{},
			ConfirmationSentAt:     time.Time{},
			UnconfirmedEmail:       "",
			Moderator:              true,
			Admin:                  true,
			Disabled:               false,
			Approved:               true,
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
		"local_account_1": {
			ID:                     "44e36b79-44a4-4bd8-91e9-097f477fe97b",
			Email:                  "zork@example.org",
			AccountID:              "580072df-4d03-4684-a412-89fd6f7d77e6",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              time.Now().Add(-36 * time.Hour),
			SignUpIP:               net.ParseIP("59.99.19.172"),
			UpdatedAt:              time.Now().Add(-72 * time.Hour),
			CurrentSignInAt:        time.Now().Add(-30 * time.Minute),
			CurrentSignInIP:        net.ParseIP("88.234.118.16"),
			LastSignInAt:           time.Now().Add(-2 * time.Hour),
			LastSignInIP:           net.ParseIP("147.111.231.154"),
			SignInCount:            9,
			InviteID:               "",
			ChosenLanguages:        []string{"en"},
			FilteredLanguages:      []string{},
			Locale:                 "en",
			CreatedByApplicationID: "",
			LastEmailedAt:          time.Now().Add(-55 * time.Minute),
			ConfirmationToken:      "",
			ConfirmedAt:            time.Now().Add(-34 * time.Hour),
			ConfirmationSentAt:     time.Now().Add(-36 * time.Hour),
			UnconfirmedEmail:       "",
			Moderator:              false,
			Admin:                  false,
			Disabled:               false,
			Approved:               true,
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
		"local_account_2": {
			ID:                     "f8d6272e-2d71-4d0c-97d3-2ba7a0b75bf7",
			Email:                  "tortle.dude@example.org",
			AccountID:              "eecaad73-5703-426d-9312-276641daa31e",
			EncryptedPassword:      "$2y$10$ggWz5QWwnx6kzb9g0tnIJurFtE0dhr5Zfeaqs9iFuUIXzafQlJVZS", // 'password'
			CreatedAt:              time.Now().Add(-36 * time.Hour),
			SignUpIP:               net.ParseIP("59.99.19.172"),
			UpdatedAt:              time.Now().Add(-72 * time.Hour),
			CurrentSignInAt:        time.Now().Add(-30 * time.Minute),
			CurrentSignInIP:        net.ParseIP("118.44.18.196"),
			LastSignInAt:           time.Now().Add(-2 * time.Hour),
			LastSignInIP:           net.ParseIP("198.98.21.15"),
			SignInCount:            9,
			InviteID:               "",
			ChosenLanguages:        []string{"en"},
			FilteredLanguages:      []string{},
			Locale:                 "en",
			CreatedByApplicationID: "",
			LastEmailedAt:          time.Now().Add(-55 * time.Minute),
			ConfirmationToken:      "",
			ConfirmedAt:            time.Now().Add(-34 * time.Hour),
			ConfirmationSentAt:     time.Now().Add(-36 * time.Hour),
			UnconfirmedEmail:       "",
			Moderator:              false,
			Admin:                  false,
			Disabled:               false,
			Approved:               true,
			ResetPasswordToken:     "",
			ResetPasswordSentAt:    time.Time{},
		},
	}

	return users
}

func TestAccounts() map[string]*model.Account {
	accounts := map[string]*model.Account{
		"unconfirmed_account": {
			ID:                    "59e197f5-87cd-4be8-ac7c-09082ccc4b4d",
			Username:              "weed_lord420",
			AvatarFileName:        "",
			AvatarContentType:     "",
			AvatarFileSize:        0,
			AvatarUpdatedAt:       time.Time{},
			AvatarRemoteURL:       "",
			HeaderFileName:        "",
			HeaderContentType:     "",
			HeaderFileSize:        0,
			HeaderUpdatedAt:       time.Time{},
			HeaderRemoteURL:       "",
			DisplayName:           "",
			Fields:                []model.Field{},
			Note:                  "",
			Memorial:              false,
			MovedToAccountID:      "",
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
			Bot:                   false,
			Reason:                "hi, please let me in! I'm looking for somewhere neato bombeato to hang out.",
			Locked:                false,
			Discoverable:          false,
			Privacy:               model.VisibilityPublic,
			Sensitive:             false,
			Language:              "en",
			URI:                   "http://localhost:8080/users/admin",
			URL:                   "http://localhost:8080/@admin",
			LastWebfingeredAt:     time.Time{},
			InboxURL:              "http://localhost:8080/users/admin/inbox",
			OutboxURL:             "http://localhost:8080/users/admin/outbox",
			SharedInboxURL:        "",
			FollowersURL:          "http://localhost:8080/users/admin/followers",
			FeaturedCollectionURL: "http://localhost:8080/users/admin/collections/featured",
			ActorType:             model.ActivityStreamsPerson,
			AlsoKnownAs:           "",
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			SensitizedAt:          time.Time{},
			SilencedAt:            time.Time{},
			SuspendedAt:           time.Time{},
			HideCollections:       false,
			SuspensionOrigin:      "",
		},
		"admin_account": {
			ID:                    "8020dbb4-1e7b-4d99-a872-4cf94e64210f",
			Username:              "admin",
			AvatarFileName:        "",
			AvatarContentType:     "",
			AvatarFileSize:        0,
			AvatarUpdatedAt:       time.Time{},
			AvatarRemoteURL:       "",
			HeaderFileName:        "",
			HeaderContentType:     "",
			HeaderFileSize:        0,
			HeaderUpdatedAt:       time.Time{},
			HeaderRemoteURL:       "",
			DisplayName:           "",
			Fields:                []model.Field{},
			Note:                  "",
			Memorial:              false,
			MovedToAccountID:      "",
			CreatedAt:             time.Now().Add(-72 * time.Hour),
			UpdatedAt:             time.Now().Add(-72 * time.Hour),
			Bot:                   false,
			Reason:                "",
			Locked:                false,
			Discoverable:          true,
			Privacy:               model.VisibilityPublic,
			Sensitive:             false,
			Language:              "en",
			URI:                   "http://localhost:8080/users/admin",
			URL:                   "http://localhost:8080/@admin",
			LastWebfingeredAt:     time.Time{},
			InboxURL:              "http://localhost:8080/users/admin/inbox",
			OutboxURL:             "http://localhost:8080/users/admin/outbox",
			SharedInboxURL:        "",
			FollowersURL:          "http://localhost:8080/users/admin/followers",
			FeaturedCollectionURL: "http://localhost:8080/users/admin/collections/featured",
			ActorType:             model.ActivityStreamsPerson,
			AlsoKnownAs:           "",
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			SensitizedAt:          time.Time{},
			SilencedAt:            time.Time{},
			SuspendedAt:           time.Time{},
			HideCollections:       false,
			SuspensionOrigin:      "",
		},
		"local_account_1": {
			ID:                    "580072df-4d03-4684-a412-89fd6f7d77e6",
			Username:              "the_mighty_zork",
			AvatarFileName:        "http://localhost:8080/fileserver/media/580072df-4d03-4684-a412-89fd6f7d77e6/avatar/original/75cfbe52-a5fb-451b-8f5a-b023229dce8d.jpeg",
			AvatarContentType:     "image/jpeg",
			AvatarFileSize:        0,
			AvatarUpdatedAt:       time.Time{},
			AvatarRemoteURL:       "",
			HeaderFileName:        "http://localhost:8080/fileserver/media/580072df-4d03-4684-a412-89fd6f7d77e6/header/original/9651c1ed-c288-4063-a95c-c7f8ff2a633f.jpeg",
			HeaderContentType:     "image/jpeg",
			HeaderFileSize:        0,
			HeaderUpdatedAt:       time.Time{},
			HeaderRemoteURL:       "",
			DisplayName:           "original zork (he/they)",
			Fields:                []model.Field{},
			Note:                  "hey yo this is my profile!",
			Memorial:              false,
			MovedToAccountID:      "",
			CreatedAt:             time.Now().Add(-48 * time.Hour),
			UpdatedAt:             time.Now().Add(-48 * time.Hour),
			Bot:                   false,
			Reason:                "I wanna be on this damned webbed site so bad! Please! Wow",
			Locked:                false,
			Discoverable:          true,
			Privacy:               model.VisibilityPublic,
			Sensitive:             false,
			Language:              "en",
			URI:                   "http://localhost:8080/users/the_mighty_zork",
			URL:                   "http://localhost:8080/@the_mighty_zork",
			LastWebfingeredAt:     time.Time{},
			InboxURL:              "http://localhost:8080/users/the_mighty_zork/inbox",
			OutboxURL:             "http://localhost:8080/users/the_mighty_zork/outbox",
			SharedInboxURL:        "",
			FollowersURL:          "http://localhost:8080/users/the_mighty_zork/followers",
			FeaturedCollectionURL: "http://localhost:8080/users/the_mighty_zork/collections/featured",
			ActorType:             model.ActivityStreamsPerson,
			AlsoKnownAs:           "",
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			SensitizedAt:          time.Time{},
			SilencedAt:            time.Time{},
			SuspendedAt:           time.Time{},
			HideCollections:       false,
			SuspensionOrigin:      "",
		},
		"local_account_2": {
			ID:                    "eecaad73-5703-426d-9312-276641daa31e",
			Username:              "1happyturtle",
			AvatarFileName:        "http://localhost:8080/fileserver/media/eecaad73-5703-426d-9312-276641daa31e/avatar/original/d5e7c265-91a6-4d84-8c27-7e1efe5720da.jpeg",
			AvatarContentType:     "image/jpeg",
			AvatarFileSize:        0,
			AvatarUpdatedAt:       time.Time{},
			AvatarRemoteURL:       "",
			HeaderFileName:        "http://localhost:8080/fileserver/media/eecaad73-5703-426d-9312-276641daa31e/header/original/e75d4117-21b6-4315-a428-eb3944235996.jpeg",
			HeaderContentType:     "image/jpeg",
			HeaderFileSize:        0,
			HeaderUpdatedAt:       time.Time{},
			HeaderRemoteURL:       "",
			DisplayName:           "happy little turtle :3",
			Fields:                []model.Field{},
			Note:                  "i post about things that concern me",
			Memorial:              false,
			MovedToAccountID:      "",
			CreatedAt:             time.Now().Add(-190 * time.Hour),
			UpdatedAt:             time.Now().Add(-36 * time.Hour),
			Bot:                   false,
			Reason:                "",
			Locked:                true,
			Discoverable:          false,
			Privacy:               model.VisibilityFollowersOnly,
			Sensitive:             false,
			Language:              "en",
			URI:                   "http://localhost:8080/users/1happyturtle",
			URL:                   "http://localhost:8080/@1happyturtle",
			LastWebfingeredAt:     time.Time{},
			InboxURL:              "http://localhost:8080/users/1happyturtle/inbox",
			OutboxURL:             "http://localhost:8080/users/1happyturtle/outbox",
			SharedInboxURL:        "",
			FollowersURL:          "http://localhost:8080/users/1happyturtle/followers",
			FeaturedCollectionURL: "http://localhost:8080/users/1happyturtle/collections/featured",
			ActorType:             model.ActivityStreamsPerson,
			AlsoKnownAs:           "",
			PrivateKey:            &rsa.PrivateKey{},
			PublicKey:             &rsa.PublicKey{},
			SensitizedAt:          time.Time{},
			SilencedAt:            time.Time{},
			SuspendedAt:           time.Time{},
			HideCollections:       false,
			SuspensionOrigin:      "",
		},
		"remote_account_1": {
			ID:       "c2c6e647-e2a9-4286-883b-e4a188186664",
			Username: "foss_satan",
			Domain:   "fossbros-anonymous.io",
		},
		"remote_account_2": {
			ID:       "93287988-76c4-460f-9e68-a45b578bb6b2",
			Username: "dailycatpics",
			Domain:   "uwu.social",
		},
		"suspended_local_account": {
			ID:       "e8a5cf4e-4b10-45a4-ad82-b6e37a09100d",
			Username: "jeffbadman",
		},
		"suspended_remote_account": {
			ID:       "17e6e09e-855d-4bf8-a1c3-7e780269f215",
			Username: "ipfreely",
			Domain:   "a-very-bad-website.com",
		},
	}

	// generate keys for each account
	for _, v := range accounts {
		priv, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			panic(err)
		}
		pub := &priv.PublicKey

		// only local accounts get a private key
		if v.Domain == "" {
			v.PrivateKey = priv
		}
		v.PublicKey = pub
	}
	return accounts
}
