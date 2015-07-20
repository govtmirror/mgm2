'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegionsCtrl
 * @description
 * # RegionsCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegionsCtrl', function($scope, $location, $timeout, $modal, mgm) {

    //keep a map of regions displayed so that we can update them quickly
    var regions = {};

    if ($scope.auth === undefined || $scope.auth === {}) {
      mgm.pushLocation($location.url());
      $location.url('/loading');
    }

    var dummyEntry = {Name: '<show all estates>', Regions: []};
    $scope.search = {
      estate: dummyEntry,
      regionName: '',
    };

    function regionsList(IDs){
      var regs = [];
      for(var i = 0; i < IDs.length; i++){
        if( IDs[i] in mgm.regions && mgm.regions[IDs[i]].Name.includes($scope.search.regionName)){
          regs.push(mgm.regions[IDs[i]]);
          regions[IDs[i]] = mgm.regions[IDs[i]];
        }
      }
      return regs;
    }

    var eMap = {};
    $scope.estates = [];
    $scope.estates.push(dummyEntry);
    for(var id in mgm.estates){
      var estate = mgm.estates[id];
      estate.regions = regionsList(estate.Regions);
      $scope.estates.push(estate);
    }

    $scope.region = {
      start: function(region) {
        console.log('Requesting start region: ' + region.Name);
        mgm.request('StartRegion', {
          RegionUUID: region.UUID
        }, function(success, msg) {
          if (success) {
            alertify.success(msg);
          } else {
            alertify.error(msg);
          }
        });
      },
      kill: function(region) {
        mgm.request('KillRegion', {
          RegionUUID: region.UUID
        }, function(success, msg) {
          if (success) {
            alertify.success(msg);
          } else {
            alertify.error(msg);
          }
        });
      },
      content: function(region) {
        alertify.error('content not implemented js: ' + region.Name);
      },
      manage: function(region) {
        if (region.Status.Running) {
          mgm.request('OpenConsole', {
            RegionUUID: region.UUID
          }, function(success, msg) {
            if (success) {
              var modInst = $modal.open({
                animation: false,
                templateUrl: 'regionConsoleModal.html',
                backdrop: 'static',
                keyboard: false,
                controller: 'RegionconsoleCtrl',
                windowClass: 'console-modal-window',
                resolve: {
                  region: function() {
                    return region;
                  }
                }
              });
              modInst.result.then(function() {
                mgm.request('CloseConsole');
              });
            } else {
              alertify.error(msg);
            }
          });
        } else {
          //prime region with estate object so forms play nicer
          var modInst = $modal.open({
            animation: false,
            templateUrl: 'manageSettingsModal.html',
            backdrop: 'static',
            keyboard: false,
            controller: 'ManageregionCtrl',
            resolve: {
              region: function() {
                return region;
              }
            }
          });
          modInst.result.then(function() {
            delete region.estate;
            delete region.host;
          });
        }
      },
      log: function(region) {
        alertify.error('log not implemented js: ' + region.Name);
      },
      showAdd: function() {
        alertify.error('Add region not implemented js');
      }
    };

    $scope.shouldShow = function(e) {
      if (e === dummyEntry) {
        return false;
      }
      if ($scope.search.estate === dummyEntry) {
        //listing all estates
        //iterate over estates, do not list estates where all regions are filtered out
        for (var i = 0; i < e.Regions.length; i++) {
          // we cannot use track by, as estate names may contain spaces
          // so angular adds $$hashKey, which we must test for
          if( e.Regions[i] === '$$hashKey') {
            continue;
          }
          if ( e.Regions[i] in mgm.regions && mgm.regions[e.Regions[i]].Name.includes($scope.search.regionName)) {
            return true;
          }
        }
        return false;
      }
      return $scope.search.estate === e;
    };

    $scope.humanReadableUptime = function(ns) {
      var seconds = ns / 1000000000;
      var days = Math.floor(seconds / 86400);
      seconds = seconds % (86400);
      var hours = Math.floor(seconds / 3600);
      seconds = seconds % (3600);
      var minutes = Math.floor(seconds / 60);
      return days + 'd ' + hours + 'h ' + minutes + 'm';
    };

    function estateifyRegion(event, region) {
      if(region.UUID in eMap) {
        $scope.estates[eMap[region.UUID]][region.UUID] = region;
      }
    }

    $scope.$on('EstateUpdate', function(event, estate){
      for (var i = 0; i < $scope.estates.length; i++) {
        if ($scope.estates[i].ID === estate.ID){
          $scope.estates[i] = estate;
        }
      }
      $timeout(function(){
        estate.regions = regionsList(estate.Regions);
      });
    });
    $scope.$on('RegionUpdate', estateifyRegion);
    $scope.$on('RegionStatusUpdate', function(event, status) {
      if(status.UUID in regions){
        $timeout(function(){
          regions[status.UUID].Status = status;
        });
      }
    });

  });
