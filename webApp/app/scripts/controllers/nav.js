'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:NavCtrl
 * @description
 * # NavCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp').controller('NavCtrl', function ($scope, $location, $timeout, mgmPublic, mgm) {

  $scope.loggedIn = false;
  $scope.$on("AuthChange", function (event, data) {
    console.log("auth changed to " + data);
    $scope.loggedIn = data;
    if ($scope.loggedIn) {
      mgm.connect();
      $timeout(function () {
        $location.url("/loading");
      }, 0.1);
    } else {
      $timeout(function () {
        $location.url("/");
      }, 0.1);
    }
  });

  $scope.isActive = function (route) {
    return $location.path().indexOf(route) == 0;
  };

  $scope.logout = function () {
    mgm.disconnect();
    mgmPublic.logout();
  };
});
