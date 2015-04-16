'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegionsCtrl
 * @description
 * # RegionsCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegionsCtrl', function ($scope, mgm) {

    var regions = {}

    $scope.estates = {}

    for (var uuid in mgm.regions) {
      estateifyRegion(mgm.regions[uuid]);
    }

    $scope.$on("RegionChange", estateifyRegion);

    function estateifyRegion(region) {
      if (region.UUID in regions) {
        //this is an update to an existing region
        angular.copy(region, regions[region.UUID]);
      } else {
        //this is a new region
        regions[region.UUID] = region;
        if (!(region.EstateName in $scope.estates)) {
          $scope.estates[region.EstateName] = [];
        }
        $scope.estates[region.EstateName].push(region)
      }
    }

  });
