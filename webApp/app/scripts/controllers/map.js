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

    if ($scope.auth === undefined || $scope.auth === {}) {
      mgm.pushLocation($location.url());
      $location.url('/loading');
    }
  });
