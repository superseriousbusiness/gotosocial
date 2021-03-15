/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package oauth

type Server struct {

}

func main() {
//    manager := manage.NewDefaultManager()
// 	// token memory store
// 	manager.MustTokenStorage(store.NewMemoryTokenStore())

// 	// client memory store
// 	clientStore := store.NewClientStore()
// 	clientStore.Set("000000", &models.Client{
// 		ID:     "000000",
// 		Secret: "999999",
// 		Domain: "http://localhost",
// 	})
// 	manager.MapClientStorage(clientStore)

// 	srv := server.NewDefaultServer(manager)
// 	srv.SetAllowGetAccessRequest(true)
// 	srv.SetClientInfoHandler(server.ClientFormHandler)

// 	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
// 		log.Println("Internal Error:", err.Error())
// 		return
// 	})

// 	srv.SetResponseErrorHandler(func(re *errors.Response) {
// 		log.Println("Response Error:", re.Error.Error())
// 	})

// 	http.HandleFunc("/authorize", func(w http.ResponseWriter, r *http.Request) {
// 		err := srv.HandleAuthorizeRequest(w, r)
// 		if err != nil {
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 		}
// 	})

// 	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
// 		srv.HandleTokenRequest(w, r)
// 	})

// 	log.Fatal(http.ListenAndServe(":9096", nil))
}
