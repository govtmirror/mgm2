'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:NavCtrl
 * @description
 * # NavCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('NavCtrl', function ($scope, $location) {
    $scope.isActive = function (route) {
      return route === $location.path();
    };
  });