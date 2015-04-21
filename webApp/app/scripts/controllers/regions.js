'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegionsCtrl
 * @description
 * # RegionsCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegionsCtrl', function ($scope, $location, mgm) {

    if ($scope.auth === undefined) {
      $location.url("/");
    }

    var regions = {}
    $scope.estates = {};

    for (var ID in mgm.estates) {
      modUserEstates("", mgm.estates[ID])
    }

    for (var uuid in mgm.regions) {
      estateifyRegion("", mgm.regions[uuid]);
    }

    $scope.$on("EstateUpdate", modUserEstates);
    $scope.$on("RegionChange", estateifyRegion);

    function estateifyRegion(event, region) {
      if (region.UUID in regions) {
        angular.copy(regions, regions[region.UUID]);
      } else {
        if (region.EstateName in $scope.estates) {
          regions[region.UUID] = region;
          $scope.estates[region.EstateName].push(region);
        }
      }
    }

    function modUserEstates(event, estate) {
      if ($scope.auth.UUID === estate.Owner || $scope.auth.UUID in estate.Managers || $scope.auth.AccessLevel > 249) {
        $scope.estates[estate.Name] = [];
      } else {
        if (estate.Name in $scope.estates) {
          delete $scope.estates[estate.Name];
          for (uuid in estate.Regions) {
            if (uuid in regions) {
              delete regions[uuid];
            }
          }
        }
      }
    }

  });
