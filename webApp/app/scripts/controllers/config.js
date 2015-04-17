'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:ConfigCtrl
 * @description
 * # ConfigCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('ConfigCtrl', function ($scope, mgm) {

    mgm.request({
      MessageType: "GetDefaultConfig"
    });

    $scope.regions = mgm.regions;
    $scope.configs = {}

    $scope.$on("ConfigUpdate", function (event, config) {
      if (!(config.Section in $scope.configs)) {
        $scope.configs[config.Section] = {}
      }
      $scope.configs[config.Section][config.Item] = config.Content;
      $scope.$apply();
    });

  });
