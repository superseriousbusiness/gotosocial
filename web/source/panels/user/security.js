/*
   GoToSocial
   Copyright (C) 2021-2022 GoToSocial Authors admin@gotosocial.org

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

"use strict";

const React = require("react");
const Promise = require("bluebird");

module.exports = function Security({oauth}) {
	const [errorMsg, setError] = React.useState("");
	const [statusMsg, setStatus] = React.useState("");

   const [oldPassword, setOldPassword] = React.useState("");
   const [newPassword, setNewPassword] = React.useState("");
   const [newPasswordConfirm, setNewPasswordConfirm] = React.useState("");

   function submit() {
		if (newPassword !== newPasswordConfirm) {
         setError("New password and confirm new password did not match!");
         return
      }
      
      setStatus("PATCHing");
		setError("");
		return Promise.try(() => {
			let formDataInfo = new FormData();
         formDataInfo.set("old_password", oldPassword);
         formDataInfo.set("new_password", newPassword);
			return oauth.apiRequest("/api/v1/user/password_change", "POST", formDataInfo, "form");
		}).then((json) => {
			setStatus("Saved!");
         setOldPassword("");
         setNewPassword("");
         setNewPasswordConfirm("");
		}).catch((e) => {
			setError(e.message);
			setStatus("");
		});
	}

   return (
      <section className="security">
            <h1>Password Change</h1>
            <form>
               <div className="labelinput">
                  <label htmlFor="oldpassword">Old password</label>
                  <input id="oldpassword" type="password" value={oldPassword} onChange={(e) => setOldPassword(e.target.value)} />
               </div>
               <div className="labelinput">
                  <label htmlFor="newpassword">New password</label>
                  <input id="newpassword" type="password" value={newPassword} onChange={(e) => setNewPassword(e.target.value)} />
               </div>
               <div className="labelinput">
                  <label htmlFor="newpasswordconfirm">Confirm new password</label>
                  <input id="newpasswordconfirm" type="password" value={newPasswordConfirm} onChange={(e) => setNewPasswordConfirm(e.target.value)} />
               </div>
            </form>
            <div className="messagebutton">
               <button onClick={submit}>Save new password</button>
               <div className="error accent">{errorMsg ? errorMsg : statusMsg}</div>
            </div>
        </section>
   )
}
