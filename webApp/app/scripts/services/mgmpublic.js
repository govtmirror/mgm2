'use strict';

/**
 * @ngdoc service
 * @name mgmApp.RegisterService
 * @description
 * # RegisterService
 * Service in the mgmApp.
 */
angular.module('mgmApp')
  .service('MgmPublic', function ($q, $http, $rootScope) {

    self = this;
    console.log("MgmPublic setting logged in to false");
    self.loggedIn = false;

    this.login = function (username, password) {
      console.log('registerUser' + username + password);

      return $q(function (resolve, reject) {

        console.log("Authenticating...");
        $http.post("/auth/login", {
          'username': username,
          'password': password
        }).success(function (data, status, headers, config) {
          if (data.Success) {
            self.loggedIn = true;
            $rootScope.$broadcast("AuthChange", true);
            resolve("login successfull");
          } else {
            reject(data.Message);
            console.log(data);
          };
        }).error(function (data, status, headers, config) {
          console.log("an error occurred");
          reject("Error connecting to MGM");
        });
      });
    };

    this.logout = function(){
      return $q(function (resolve, reject) {

        console.log("Authenticating...");
        $http.get("/auth/logout").success(function (data, status, headers, config) {
          if (data.Success) {
            self.loggedIn = false;
            $rootScope.$broadcast("AuthChange", false);
            resolve("login successfull");
          } else {
            reject(data.Message);
            console.log(data);
          };
        }).error(function (data, status, headers, config) {
          console.log("an error occurred");
          reject("Error connecting to MGM");
        });
      });
    }

    this.registerUser = function (uname, email, gender, pword, reason) {

      console.log('registerUser' + uname + email + gender + pword + reason);

      return $q(function (resolve, reject) {
        reject('Not Implemented');
      });



      /*
      for (var i = 0; i < users.length; i++) {
        if (users[i].name == uname) {
          $scope.error.other = 'User ' + uname + ' already exists on the grid';
          return;
        }
        if (users[i].email == email) {
          $scope.error.other = 'Email ' + email + 'already exists on the grid');
        return;
      }
      */
    };

    this.resumeSession = function () {

    }

    this.requestResetCode = function (email) {
      console.log('reset requested for ' + email);

      return $q(function (resolve, reject) {
        reject('Not Implemented');
      });
    };

    this.resetPassword = function (uname, token, password) {
      console.log('registerUser' + uname + token + password);

      return $q(function (resolve, reject) {
        reject('Not Implemented');
      });
    };

    console.log("resuming session...");
    //resume session functionality
    $http.get("/auth").success(function (data, status, headers, config) {
      console.log("resume session received: ");
      console.log(data);
      if (data.Success) {
        console.log("session resume successfull");
        self.loggedIn = true;
        $rootScope.$broadcast("AuthChange", true);
      };
    });
  });