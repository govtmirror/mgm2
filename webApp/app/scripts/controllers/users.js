'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:UsersCtrl
 * @description
 * # UsersCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('UsersCtrl', function ($scope, $routeParams, mgm) {

    $scope.section = $routeParams["section"];
    if (!$scope.section) {
      $scope.section = "active";
    }

    $scope.isActive = function (section) {
      return this.section === section;
    }

    $scope.users = mgm.users;
    $scope.$on("UserUpdate", function (event, user) {
      if (!(user.UserID in $scope.users)) {
        $scope.users[user.UserID] = user;
      } else {
        angular.copy(user, $scope.users[user.UserID]);
      }
    });
  });
