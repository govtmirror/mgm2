'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:ConfigCtrl
 * @description
 * # ConfigCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('ConfigCtrl', function ($scope, $routeParams, $location, $timeout, mgm) {

    if ($scope.auth === undefined || $scope.auth === {}) {
      mgm.pushLocation($location.url());
      $location.url('/loading');
    }

    $scope.regions = [];
    $scope.estates = mgm.estates;
    $scope.currentEstate = undefined;
    $scope.currentRegion = {
      'UUID': '00000000-0000-0000-0000-000000000000'
    };
    $scope.defaultConfig = {};
    $scope.regionConfig = {};
    $scope.editConfig = {};

    //list regions when estate is selected
    $scope.displayEstate = function () {
      populateRegions();
      if ($routeParams.estate !== $scope.currentEstate.ID) {
        console.log('redirecting to estate');
        $location.url('/config/' + $scope.currentEstate.ID);
      }
    };

    function populateRegions() {
      var regions = [];
      if ($scope.currentEstate !== undefined) {
        for (var i = 0; i < $scope.currentEstate.Regions.length; i++) {
          regions.push(mgm.regions[$scope.currentEstate.Regions[i]]);
        }
      }
      $scope.regions = regions;
    };

    function requestConfig(uuid) {
      $scope.config = {};
      var region = '';
      if (uuid === '00000000-0000-0000-0000-000000000000') {
        console.log('Requesting default configs');
      } else {
        console.log('Requesting configs for ' + uuid);
        region = uuid;
      }
      mgm.ws.emit('GetConfig', region, function (response) {
        response = angular.fromJson(response);
        if (response.Success === true) {
          for (var i = 0; i < response.Configs.length; i++) {
            processConfig(response.Configs[i]);
          }
          $timeout(function () {
            generateEditConfig();
          });
        } else {
          alertify.error(response.Message);
        }
      });
    };

    function processConfig(cfg) {
      if (cfg.Region === '00000000-0000-0000-0000-000000000000') {
        if (!(cfg.Section in $scope.defaultConfig)) {
          $scope.defaultConfig[cfg.Section] = {};
        }
        $scope.defaultConfig[cfg.Section][cfg.Item] = cfg.Content;
      }
      if (cfg.Region === $scope.currentRegion.UUID) {
        if (!(cfg.Section in $scope.regionConfig)) {
          $scope.regionConfig[cfg.Section] = {};
        }
        $scope.regionConfig[cfg.Section][cfg.Item] = cfg.Content;
      }
    };

    $scope.displayConfig = function () {
      if ($scope.currentEstate !== undefined) {
        if ($scope.currentRegion !== undefined) {
          $location.url('/config/' + $scope.currentEstate.ID + '/' + $scope.currentRegion.UUID);
        } else {
          $location.url('/config/' + $scope.currentEstate.ID);
        }
      } else {
        $location.url('/config');
      }
    };

    //assign variables from url, where possible
    if ($routeParams.estate !== undefined) {
      $scope.currentEstate = mgm.estates[$routeParams.estate];
      populateRegions();
    }
    if ($routeParams.region !== undefined) {
      $scope.currentRegion = mgm.regions[$routeParams.region];
      requestConfig($routeParams.region);
    }

    function generateEditConfig() {
      var newConfig = {};
      //populate default options
      angular.forEach($scope.defaultConfig, function (row, section) {
        if (newConfig[section] === undefined) {
          newConfig[section] = {};
        }
        angular.forEach(row, function (value, key) {
          newConfig[section][key] = {
            'value': value,
            'source': 'default'
          };
        });
      });
      //insert region specific options, overwriting is by design
      angular.forEach($scope.regionConfig, function (row, section) {
        if (newConfig[section] === undefined) {
          newConfig[section] = {};
        }
        angular.forEach(row, function (value, key) {
          newConfig[section][key] = {
            'value': value,
            'source': 'region'
          };
        });
      });
      $scope.editConfig = newConfig;
    };

    requestConfig('00000000-0000-0000-0000-000000000000');
  });
