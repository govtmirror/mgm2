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

    this.logout = function () {
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

    this.registerUser = function (uname, email, template, password, reason) {
      return $q(function (resolve, reject) {
        $http.post("/auth/register", {
          "Name": uname,
          "Email": email,
          "Template": template,
          "Password": password,
          "Summary": reason,
        }).success(function (data, status, headers, config) {
          if (data.Success) {
            resolve("Registration Successful");
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

    this.requestResetCode = function (email) {
      return $q(function (resolve, reject) {
        $http.post("/auth/passwordToken", {
          "Email": email
        }).success(function (data, status, headers, config) {
          if (data.Success) {
            resolve("reset code successfully requested");
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

    this.resetPassword = function (uname, token, password) {
      return $q(function (resolve, reject) {
        $http.post("/auth/passwordReset", {
          "Name": uname,
          "Token": token,
          "Password": password
        }).success(function (data, status, headers, config) {
          if (data.Success) {
            resolve("Password successfully changed");
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
