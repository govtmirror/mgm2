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

    if ($scope.auth === undefined) {
      $location.url("/login");
    }

    $scope.account = {
      UserID: "",
      Name: "",
      AccessLevel: "",
      Email: ""
    };

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



  })
