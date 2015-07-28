'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:UsersCtrl
 * @description
 * # UsersCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('UsersCtrl', function ($scope, $location, $routeParams, mgm) {

    if ($scope.auth === undefined || $scope.auth === {}) {
      mgm.pushLocation($location.url());
      $location.url('/loading');
    }

    $scope.section = $routeParams.section;
    if (!$scope.section) {
      $scope.section = 'active';
    }

    $scope.isActive = function (section) {
      return this.section === section;
    };

    var users = {}

    $scope.Math = window.Math;

    $scope.activeUsers = [];
    for (var id in mgm.activeUsers) {
      var user = mgm.activeUsers[id];
      $scope.activeUsers.push(user);
      users[id] = user;
    }

    $scope.suspendedUsers = [];
    for( var id in mgm.suspendedUsers){
      var user = mgm.suspendedUsers[id];
      $scope.suspendedUsers.push(user);
      users[id] = user;
    }
    $scope.$on('UserUpdate', function (event, user) {
      //if (user.Suspended) {
      //  console.log(user);
      //}
      //if (!(user.UserID in $scope.users)) {
      //  $scope.users[user.UserID] = user;
      //} else {
      //  angular.copy(user, $scope.users[user.UserID]);
      //}
    });

    $scope.pendingUsers = [];
    for( var id in mgm.pendingUsers){
      var user = mgm.pendingUsers[id];
      $scope.pendingUsers.push(user);
      users[id] = user;
    }
    $scope.$on('UserUpdate', function (event, user) {
      if (user.Suspended) {
        if (user.UserID in $scope.activeUsers) {
          delete $scope.activeUsers[user.UserID];
          $scope.suspendedUsers[user.UserID] = user;
        } else {
          angular.copy(user, $scope.suspendedUsers[user.UserID]);
        }
      } else {
        if (user.UserID in $scope.suspendedUsers) {
          delete $scope.suspendedUsers[user.UserID];
          $scope.activeUsers[user.UserID] = user;
        } else {
          angular.copy(user, $scope.activeUsers[user.UserID]);
        }
      }
    });
  });
