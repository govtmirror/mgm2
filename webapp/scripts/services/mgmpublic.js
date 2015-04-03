'use strict';

/**
 * @ngdoc service
 * @name mgmApp.RegisterService
 * @description
 * # RegisterService
 * Service in the mgmApp.
 */
angular.module('mgmApp')
  .service('MgmPublic', function ($q, $http) {

    this.login = function (username, password) {
      console.log('registerUser' + username + password);

      return $q(function (resolve, reject) {

        console.log("Authenticating...");
        $http.post("/auth", {
          'username': username,
          'password': password
        }).success(function (data, status, headers, config) {
          console.log("success function");
          if (data.Success) {
            reject("login successfull");
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
  });