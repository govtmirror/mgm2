'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:AccountCtrl
 * @description
 * # AccountCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('AccountCtrl', function ($scope, $location, $timeout, mgm) {

    if ($scope.auth === undefined || $scope.auth === {}) {
      $location.url("/loading");
    }

    $scope.account = {
      UserID: '',
      Name: '',
      AccessLevel: '',
      Email: ''
    };

    $scope.password = {
      passwordError: '',
      disablePasswordSubmit: false,
      passwordResult: '',
      password: '',
      confirm: ''
    }

    $scope.nav = {
      state: '',
      toggleLoad: function () {
        if ($scope.nav.state === 'load') {
          $scope.nav.state = '';
        } else {
          $scope.nav.state = 'load';
        }
      },
      toggleSave: function () {
        if ($scope.nav.state === 'save') {
          $scope.nav.state = '';
        } else {
          $scope.nav.state = 'save';
        }
      },
      togglePass: function () {
        if ($scope.nav.state === 'pass') {
          $scope.nav.state = '';
        } else {
          $scope.nav.state = 'pass';
        }
      }
    }

    for (var uuid in mgm.activeUsers) {
      if (uuid === $scope.auth.UUID) {
        angular.copy(mgm.activeUsers[uuid], $scope.account);
      }
    }

    $scope.$on("UserUpdate", function (event, user) {
      if (user.UserID === $scope.auth.UUID) {
        angular.copy(user, $scope.account);
        $scope.$apply();
      };
    });

    $scope.setPassword = function () {
      $scope.password.passwordError = '';
      if ($scope.password.password === undefined || $scope.password.password === '') {
        $scope.password.passwordError = 'Password cannot be blank';
        return;
      }
      if ($scope.password.confirm === undefined || $scope.password.confirm === '') {
        $scope.password.passwordError = 'Password confirmation is blank';
        return;
      }
      if ($scope.password.confirm !== $scope.password.password) {
        $scope.password.passwordError = 'Passwords do not match';
        return;
      }
      $scope.password.disablePasswordSubmit = true;
      mgm.request("SetPassword", {
        UserID: $scope.auth.UUID,
        Password: $scope.password.password
      }, function (success, message) {
        if (success === true) {
          $scope.password.passwordResult = 'password updated successfuly';
          $scope.password.password = '';
          $scope.password.confirm = '';
          $timeout(function () {
            $scope.password.passwordResult = '';
          }, 5 * 1000);
        } else {
          $scope.password.passwordError = message;
        }
        $scope.password.disablePasswordSubmit = false;
        $scope.$apply();
      });
    }

    $scope.iar = {
      password: '',
      file: undefined,
      message: '',
      upload: function () {
        if ($scope.iar.file === undefined) {
          console.log("invalid file selection");
          console.log($scope.iar.file);
          return;
        }
        $scope.iar.message = 'Uploading...';
        //request iar upload from mgm
        console.log("requesting from mgm");
        mgm.request("IarUpload", {
          UserID: $scope.auth.UUID,
          Password: $scope.iar.password
        }, function (success, message) {
          console.log("success from mgm request: " + success + " message: " + message)
          if (success === true) {
            mgm.upload("/upload/" + message, $scope.iar.file[0]).then(
              function () {
                //success
                console.log("upload must have worked");
                $scope.iar.password = '';
                $scope.iar.file = undefined;
                $scope.iar.message = '';
              },
              function (msg) {
                //error
                console.log(msg);
                $scope.iar.message = 'Error: ' + msg;
              }
            );
          } else {
            $scope.iar.message = message;
          }
          $scope.$apply();
        });
      }
    }

  });
