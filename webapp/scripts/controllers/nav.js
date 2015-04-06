'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:NavCtrl
 * @description
 * # NavCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp').controller('NavCtrl', function ($scope, $location, MgmPublic) {
  
  $scope.loggedIn = false;
  $scope.$on("AuthChange", function(){
    $scope.loggedIn = MgmPublic.loggedIn;
  });
  
  $scope.isActive = function (route) {
    return route === $location.path();
  };

  $scope.logout = function(){
    MgmPublic.logout();
  };
});