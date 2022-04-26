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

package log_test

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
	"github.com/superseriousbusiness/gotosocial/internal/config"
	"github.com/superseriousbusiness/gotosocial/testrig"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

type SyslogTestSuite struct {
	suite.Suite
	syslogServer  *syslog.Server
	syslogChannel chan format.LogParts
}

func (suite *SyslogTestSuite) SetupTest() {
	testrig.InitTestConfig()

	viper.Set(config.Keys.SyslogEnabled, true)
	viper.Set(config.Keys.SyslogProtocol, "udp")
	viper.Set(config.Keys.SyslogAddress, "localhost:42069")
	server, channel, err := testrig.InitTestSyslog()
	if err != nil {
		panic(err)
	}
	suite.syslogServer = server
	suite.syslogChannel = channel

	testrig.InitTestLog()
}

func (suite *SyslogTestSuite) TearDownTest() {
	if err := suite.syslogServer.Kill(); err != nil {
		panic(err)
	}
}

func (suite *SyslogTestSuite) TestSyslog() {
	logrus.Warn("this is a test of the emergency broadcast system!")

	message := <-suite.syslogChannel
	suite.Contains(message["content"], "msg=this is a test of the emergency broadcast system!")
}

func (suite *SyslogTestSuite) TestSyslogLongMessage() {
	longMessage := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Elit ut aliquam purus sit. Fringilla est ullamcorper eget nulla facilisi. Nisi vitae suscipit tellus mauris a diam maecenas sed. Sed adipiscing diam donec adipiscing tristique risus nec feugiat. Consequat id porta nibh venenatis cras sed felis eget velit. Dignissim sodales ut eu sem integer vitae justo eget magna. Pellentesque habitant morbi tristique senectus et netus et malesuada. Semper eget duis at tellus at urna condimentum mattis. Ac auctor augue mauris augue. Erat velit scelerisque in dictum non consectetur a erat nam. Accumsan in nisl nisi scelerisque eu ultrices. Amet justo donec enim diam vulputate ut pharetra sit amet. Elit pellentesque habitant morbi tristique senectus et. Elementum nibh tellus molestie nunc. Phasellus vestibulum lorem sed risus ultricies tristique nulla aliquet. Dictum fusce ut placerat orci nulla pellentesque dignissim. Vel elit scelerisque mauris pellentesque. Dignissim cras tincidunt lobortis feugiat vivamus. Massa enim nec dui nunc mattis. Sed sed risus pretium quam vulputate dignissim suspendisse. Purus viverra accumsan in nisl nisi scelerisque eu ultrices vitae. Vulputate enim nulla aliquet porttitor lacus. Sed lectus vestibulum mattis ullamcorper. Lorem ipsum dolor sit amet. Egestas diam in arcu cursus euismod quis viverra. Varius sit amet mattis vulputate enim nulla. Proin sagittis nisl rhoncus mattis rhoncus urna neque. Arcu non odio euismod lacinia at quis risus. Nulla facilisi cras fermentum odio eu feugiat pretium. Molestie a iaculis at erat pellentesque adipiscing commodo elit. Etiam tempor orci eu lobortis. Eget nulla facilisi etiam dignissim diam. Scelerisque mauris pellentesque pulvinar pellentesque habitant morbi tristique senectus et. Commodo quis imperdiet massa tincidunt nunc pulvinar. Ullamcorper morbi tincidunt ornare massa eget egestas purus. Sollicitudin aliquam ultrices sagittis orci a scelerisque purus semper. Vel facilisis volutpat est velit egestas. Diam sit amet nisl suscipit adipiscing bibendum est. Phasellus egestas tellus rutrum tellus. Eu ultrices vitae auctor eu augue ut lectus. Lorem sed risus ultricies tristique nulla aliquet enim. Eget nunc scelerisque viverra mauris in aliquam sem. Lorem ipsum dolor sit amet. Sit amet porttitor eget dolor morbi. Lacus suspendisse faucibus interdum posuere lorem ipsum. Nec ullamcorper sit amet risus nullam eget felis eget. Dignissim convallis aenean et tortor. Purus in mollis nunc sed id semper risus in. Amet aliquam id diam maecenas ultricies mi. Orci phasellus egestas tellus rutrum tellus. Tristique sollicitudin nibh sit amet commodo. Aliquet bibendum enim facilisis gravida. Morbi tempus iaculis urna id volutpat. Integer eget aliquet nibh praesent tristique magna sit amet purus. Eu augue ut lectus arcu. Rhoncus dolor purus non enim praesent elementum facilisis. A condimentum vitae sapien pellentesque habitant morbi tristique. Aliquet porttitor lacus luctus accumsan tortor posuere ac ut. Facilisis mauris sit amet massa vitae tortor condimentum lacinia quis. Morbi non arcu risus quis varius quam quisque. Metus aliquam eleifend mi in nulla posuere sollicitudin aliquam. Neque volutpat ac tincidunt vitae semper quis lectus nulla at. Vestibulum sed arcu non odio euismod lacinia at quis. Aenean sed adipiscing diam donec. Consequat ac felis donec et odio pellentesque diam. Placerat orci nulla pellentesque dignissim enim sit amet. Tempor commodo ullamcorper a lacus vestibulum sed arcu. Mollis aliquam ut porttitor leo a diam sollicitudin tempor. Aliquet risus feugiat in ante metus dictum at tempor commodo. Enim nulla aliquet porttitor lacus luctus accumsan tortor posuere. Mattis nunc sed blandit libero. Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Elit ut aliquam purus sit. Fringilla est ullamcorper eget nulla facilisi. Nisi vitae suscipit tellus mauris a diam maecenas sed. Sed adipiscing diam donec adipiscing tristique risus nec feugiat. Consequat id porta nibh venenatis cras sed felis eget velit. Dignissim sodales ut eu sem integer vitae justo eget magna. Pellentesque habitant morbi tristique senectus et netus et malesuada. Semper eget duis at tellus at urna condimentum mattis. Ac auctor augue mauris augue. Erat velit scelerisque in dictum non consectetur a erat nam. Accumsan in nisl nisi scelerisque eu ultrices. Amet justo donec enim diam vulputate ut pharetra sit amet. Elit pellentesque habitant morbi tristique senectus et. Elementum nibh tellus molestie nunc. Phasellus vestibulum lorem sed risus ultricies tristique nulla aliquet. Dictum fusce ut placerat orci nulla pellentesque dignissim. Vel elit scelerisque mauris pellentesque. Dignissim cras tincidunt lobortis feugiat vivamus. Massa enim nec dui nunc mattis. Sed sed risus pretium quam vulputate dignissim suspendisse. Purus viverra accumsan in nisl nisi scelerisque eu ultrices vitae. Vulputate enim nulla aliquet porttitor lacus. Sed lectus vestibulum mattis ullamcorper. Lorem ipsum dolor sit amet. Egestas diam in arcu cursus euismod quis viverra. Varius sit amet mattis vulputate enim nulla. Proin sagittis nisl rhoncus mattis rhoncus urna neque. Arcu non odio euismod lacinia at quis risus. Nulla facilisi cras fermentum odio eu feugiat pretium. Molestie a iaculis at erat pellentesque adipiscing commodo elit. Etiam tempor orci eu lobortis. Eget nulla facilisi etiam dignissim diam. Scelerisque mauris pellentesque pulvinar pellentesque habitant morbi tristique senectus et. Commodo quis imperdiet massa tincidunt nunc pulvinar. Ullamcorper morbi tincidunt ornare massa eget egestas purus. Sollicitudin aliquam ultrices sagittis orci a scelerisque purus semper. Vel facilisis volutpat est velit egestas. Diam sit amet nisl suscipit adipiscing bibendum est. Phasellus egestas tellus rutrum tellus. Eu ultrices vitae auctor eu augue ut lectus. Lorem sed risus ultricies tristique nulla aliquet enim. Eget nunc scelerisque viverra mauris in aliquam sem. Lorem ipsum dolor sit amet. Sit amet porttitor eget dolor morbi. Lacus suspendisse faucibus interdum posuere lorem ipsum. Nec ullamcorper sit amet risus nullam eget felis eget. Dignissim convallis aenean et tortor. Purus in mollis nunc sed id semper risus in. Amet aliquam id diam maecenas ultricies mi. Orci phasellus egestas tellus rutrum tellus. Tristique sollicitudin nibh sit amet commodo. Aliquet bibendum enim facilisis gravida. Morbi tempus iaculis urna id volutpat. Integer eget aliquet nibh praesent tristique magna sit amet purus. Eu augue ut lectus arcu. Rhoncus dolor purus non enim praesent elementum facilisis. A condimentum vitae sapien pellentesque habitant morbi tristique. Aliquet porttitor lacus luctus accumsan tortor posuere ac ut. Facilisis mauris sit amet massa vitae tortor condimentum lacinia quis. Morbi non arcu risus quis varius quam quisque. Metus aliquam eleifend mi in nulla posuere sollicitudin aliquam. Neque volutpat ac tincidunt vitae semper quis lectus nulla at. Vestibulum sed arcu non odio euismod lacinia at quis. Aenean sed adipiscing diam donec. Consequat ac felis donec et odio pellentesque diam. Placerat orci nulla pellentesque dignissim enim sit amet. Tempor commodo ullamcorper a lacus vestibulum sed arcu. Mollis aliquam ut porttitor leo a diam sollicitudin tempor. Aliquet risus feugiat in ante metus dictum at tempor commodo. Enim nulla aliquet porttitor lacus luctus accumsan tortor posuere. Mattis nunc sed blandit libero."
	
	logrus.Warn(longMessage)

	message := <-suite.syslogChannel
	suite.Contains(message["content"], "this is a test of the emergency broadcast system!")
}

func TestSyslogTestSuite(t *testing.T) {
	suite.Run(t, &SyslogTestSuite{})
}
