'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegionconsoleCtrl
 * @description
 * # RegionconsoleCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegionconsoleCtrl', function ($scope, $modalInstance, region) {

    $scope.region = region;

    $scope.close = function() {
      $modalInstance.close();
    };

  });
