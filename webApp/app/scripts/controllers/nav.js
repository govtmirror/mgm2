'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:NavCtrl
 * @description
 * # NavCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp').controller('NavCtrl', function ($scope, $location, MgmPublic, mgm) {

  $scope.loggedIn = false;
  $scope.$on("AuthChange", function (event, data) {
    $scope.loggedIn = data;
    if ($scope.loggedIn) {
      $location.url("/account");
      console.log("redirecting to Account page");
      mgm.connect();
    } else {
      $location.url("/");
      console.log("redirecting to Home page");
    }
  });

  $scope.isActive = function (route) {
    return route === $location.path();
  };

  $scope.logout = function () {
    mgm.disconnect();
    MgmPublic.logout();
  };
});
