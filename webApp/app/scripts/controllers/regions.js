'use strict';

/**
 * @ngdoc function
 * @name mgmApp.controller:RegionsCtrl
 * @description
 * # RegionsCtrl
 * Controller of the mgmApp
 */
angular.module('mgmApp')
  .controller('RegionsCtrl', function ($scope, $location, $timeout, Alertify, mgm) {

    if ($scope.auth === undefined || $scope.auth === {}) {
      mgm.pushLocation($location.url());
      $location.url("/loading");
    }

    var dummyEntry = ' <show all estates>';

    var regions = {}
    $scope.estates = {};
    $scope.estates[dummyEntry] = {};
    $scope.search = {
      estateName: dummyEntry,
      regionName: '',
    };

    $scope.region = {
      start: function(region){
        console.log("Requesting start region: " + region.Name);
        mgm.request("StartRegion", {RegionUUID: region.UUID}, function(success, msg){
          if(success) {
            Alertify.success(msg);
          } else {
            Alertify.error(msg);
          }
        })
      },
      kill: function(region){
        mgm.request("KillRegion", {RegionUUID: region.UUID}, function(success, msg){
          if(success) {
            Alertify.success(msg);
          } else {
            Alertify.error(msg);
          }
        });
      },
      content: function(region){
        Alertify.error("content not implemented js");
      },
      manage: function(region){
        Alertify.error("manage not implemented js");
      },
      log: function(region){
        Alertify.error("log not implemented js");
      },
      showAdd: function(){
        Alertify.error("Add region not implemented js");
      }
    }

    $scope.shouldShow = function(name){
      if( name === dummyEntry)
        return false;
      if($scope.search.estateName === dummyEntry){
        //listing all estates
        //iterate over estates, do not list estates where all regions are filtered out
        for(var uuid in $scope.estates[name]){
          if($scope.estates[name][uuid].Name.includes($scope.search.regionName)){
            return true;
          }
        }
        return false;
      }
      return $scope.search.estateName === name;
    }

    $scope.humanReadableUptime = function(ns){
      var seconds = ns / 1000000000;
      var days = Math.floor(seconds/86400)
      seconds = seconds % (86400)
      var hours = Math.floor(seconds/3600)
      seconds = seconds % (3600)
      var minutes = Math.floor(seconds/60)
      return days+"d "+hours+"h "+minutes+"m";
    }


    for (var ID in mgm.estates) {
      modUserEstates("", mgm.estates[ID])
    }

    for (var uuid in mgm.regions) {
      estateifyRegion("", mgm.regions[uuid]);
    }

    $scope.$on("EstateUpdate", modUserEstates);
    $scope.$on("RegionUpdate", estateifyRegion);
    $scope.$on("RegionStatusUpdate", function(event, status){
      if(status.UUID in regions){
        $timeout(function(){
          $scope.estates[regions[status.UUID].EstateName][status.UUID].Status = status
        })
      }
    })

    function estateifyRegion(event, region) {
      if (region.UUID in regions) {
        regions[region.UUID] = region;
      } else {
        if (region.EstateName in $scope.estates) {
          regions[region.UUID] = region;
          $scope.estates[region.EstateName][region.UUID] = region;
        } else {
          $scope.estates[region.EstateName] = {};
          regions[region.UUID] = region;
          $scope.estates[region.EstateName][region.UUID] = region;
        }
      }
    }

    function modUserEstates(event, estate) {
      if ($scope.auth.UUID === estate.Owner || $scope.auth.UUID in estate.Managers || $scope.auth.AccessLevel > 249) {
        if( !estate.Name in $scope.estates){
          $scope.estates[estate.Name] = {};
        }
      } else {
        //remove estate, this user no longer controlls it
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
