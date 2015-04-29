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
        $scope.passwordError = 'Password cannot be blank';
        return;
      }
      if ($scope.password.confirm === undefined || $scope.password.confirm === '') {
        $scope.password.passwordError = 'Password confirmation cannot be blank';
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

  });
