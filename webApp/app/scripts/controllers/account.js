'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:AccountCtrl
 * @description
 * # AccountCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('AccountCtrl', function ($scope, mgm) {

    $scope.account = {
      UserID: "",
      Name: "",
      AccessLevel: "",
      Email: ""
    };

    for (var uuid in mgm.users) {
      if (uuid === $scope.auth.UUID) {
        angular.copy(mgm.users[uuid], $scope.account);
      }
    }

    $scope.$on("UserUpdate", function (event, user) {
      if (user.UserID === $scope.auth.UUID) {
        //$scope.account = user;
        //$scope.$apply();
        console.log("Applying user to account");
        angular.copy(user, $scope.account);
        $scope.$apply();
      };
    });



  })
