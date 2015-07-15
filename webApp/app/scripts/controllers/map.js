'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:MapCtrl
 * @description
 * # MapCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('MapCtrl', function ($scope, $location, mgm) {

    $scope.regions = mgm.regions;
    $scope.centerX = 1000;
    $scope.centerY = 1000;

    if ($scope.auth === undefined || $scope.auth === {}) {
      mgm.pushLocation($location.url());
      $location.url('/loading');
    }
  });
