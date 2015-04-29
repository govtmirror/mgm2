'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:AccountCtrl
 * @description
 * # AccountCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('AccountCtrl', function ($scope, $location, mgm) {

    if ($scope.auth === undefined || $scope.auth === {}) {
      $location.url("/loading");
    }

    $scope.account = {
      UserID: '',
      Name: '',
      AccessLevel: '',
      Email: ''
    };

    $scope.passwordError = '';
    $scope.disablePasswordSubmit = false;

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

    $scope.setPassword = function (password, confirm) {
      $scope.passwordError = '';
      if (password === undefined || password === '') {
        $scope.passwordError = 'Password cannot be blank';
        return;
      }
      if (confirm === undefined || confirm === '') {
        $scope.passwordError = 'Password confirmation cannot be blank';
        return;
      }
      if (confirm !== password) {
        $scope.passwordError = 'Passwords do not match';
        return;
      }
      mgm.request("SetPassword", {
        UserID: $scope.auth.UUID,
        Password: password
      }, function (success, message) {
        $scope.passwordError = message;
        $scope.$apply();
      });


    }

  });
