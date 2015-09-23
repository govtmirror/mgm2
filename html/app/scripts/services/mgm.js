'use strict';

/**
 * @ngdoc service
 * @name mgmApp.mgm
 * @description
 * # mgm
 * Service in the mgmApp.
 */
angular.module('mgmApp').service('mgm', function ($location, $rootScope, $q, $http) {
  console.log('mgm service instantiated');

  var remoteURL = 'ws://' + $location.host() + ':' + $location.port() + '/ws';

  var self = this;

  self.regions = {};
  self.estates = {};
  self.activeUsers = {};
  self.suspendedUsers = {};
  self.pendingUsers = {};
  self.groups = {};
  self.hosts = {};
  self.jobs = {};
  self.serverConnected = false;

  $rootScope.$on('AuthChange', function (event, value) {
    if (value === false) {
      //we logged out
      self.regions = {};
      self.estates = {};
      self.activeUsers = {};
      self.suspendedUsers = {};
      self.pendingUsers = {};
      self.groups = {};
      self.hosts = {};
      self.serverConnected = false;
    }
  });

  self.connect = function () {
    $rootScope.$broadcast('SyncBegin');
    console.log('Connecting to: ' + remoteURL);

    self.ws = io({
      path: '/ws',
      'force new connection': true
    });

    self.ws.on('Auth Challenge', function () {
      console.log('Auth challenge received, sending token');
      self.ws.emit('Auth Response', self.token, function () {
        console.log('token accepted');

        $rootScope.$broadcast('ServerConnected');
        self.serverConnected = true;
        self.ws.emit('GetState', '', function (data) {
          console.log('state received');
          data = angular.fromJson(data);
          //consume state data
          for(var i = 0; i < data.Users.length; i++){
            onUser(data.Users[i]);
          }
          for(var i = 0; i < data.PendingUsers.length; i++){
            onPendingUser(data.PendingUsers[i]);
          }
          for(var i = 0; i < data.Estates.length; i++){
            onEstate(data.Estates[i]);
          }
          for(var i = 0; i < data.Groups.length; i++){
            onGroup(data.Groups[i]);
          }
          for(var i = 0; i < data.Jobs.length; i++){
            onJob(data.Jobs[i]);
          }
          for(var i = 0; i < data.Hosts.length; i++){
            onHost(data.Hosts[i]);
          }
          for(var i = 0; i < data.HostStats.length; i++){
            onHostStat(data.HostStats[i]);
          }
          for(var i = 0; i < data.Regions.length; i++){
            onRegion(data.Regions[i]);
          }
          for(var i = 0; i < data.RegionStats.length; i++){
            onRegionStat(data.RegionStats[i]);
          }

          $rootScope.$broadcast('SyncComplete');
        });
      });
    });

    self.ws.on('error', function(err){
      console.log(err);
    })

    self.ws.on('connect', function () {
      console.log('Socket has been opened!');
    });

    self.ws.on('disconnect', function () {
      console.log('socket has been closed');
    })

    self.ws.on('User', onUser);
    self.ws.on('PendingUser', onPendingUser);
    self.ws.on('Estate', onEstate);
    self.ws.on('Group', onGroup);
    self.ws.on('Job', onJob);
    self.ws.on('Host', onHost);
    self.ws.on('HostRemoved', onHostRemoved);
    self.ws.on('HostStat', onHostStat);
    self.ws.on('Region', onRegion);

    function onUser(data){
      var user = angular.fromJson(data);
      if (user.Suspended) {
        self.suspendedUsers[user.UserID] = user;
        if (user.UserID in self.activeUsers) {
          delete self.activeUsers[user.UserID];
        }
      } else {
        self.activeUsers[user.UserID] = user;
        if (user.UserID in self.suspendedUsers) {
          delete self.suspendedUsers[user.UserID];
        }
      }
      $rootScope.$broadcast('UserUpdate', user);
    }

    function onPendingUser(data){
      var user = angular.fromJson(data);
      self.pendingUsers[user.UserID] = user;
      $rootScope.$broadcast('PendingUserUpdate', user);
    }

    function onRegion(data){
      var region = angular.fromJson(data);
      self.regions[region.UUID] = region;
      $rootScope.$broadcast('RegionUpdate', region);
    }

    function onRegionStat(data){
      var rStat = angular.fromJson(data);
      if (rStat.UUID in self.regions) {
        self.regions[rStat.UUID].Status = rStat;
        $rootScope.$broadcast('RegionStatusUpdate', rStat);
      }
    }

    function onEstate(data){
      var estate = angular.fromJson(data);
      self.estates[estate.ID] = estate;
      $rootScope.$broadcast('EstateUpdate', estate);
    }

    function onGroup(data){
      var group = angular.fromJson(data);
      self.groups[group.ID] = group;
      $rootScope.$broadcast('GroupUpdate', group);
    }

    function onJob(data){
      var job = angular.fromJson(data);
      job.Data = angular.fromJson(job.Data);
      self.jobs[job.ID] = job;
      $rootScope.$broadcast('JobUpdate', job);
    }

    function onHost(data){
      var host = angular.fromJson(data);
      self.hosts[host.ID] = host;
      $rootScope.$broadcast('HostUpdate', host);
    }

    function onHostRemoved(data){
      delete self.hosts[data];
      $rootScope.$broadcast('HostRemoved', data);
    }

    function onHostStat(data){
      var hStat = angular.fromJson(data);
      if (hStat.ID in self.hosts) {
        self.hosts[hStat.ID].Status = hStat;
        $rootScope.$broadcast('HostStatusUpdate', hStat);
      }
    }
    /*self.ws.onmessage = function (evt) {
      var message = angular.fromJson(evt.data);
      switch (message.MessageType) {
      case 'RegionDeleted':
        delete self.regions[message.Message.UUID];
        $rootScope.$broadcast('RegionDeleted', message.Message);
        break;
      case 'JobDeleted':
        delete self.jobs[message.Message.ID]
        $rootScope.$broadcast('JobDeleted', message.Message)
        break;
      case 'RegionConsole':
        if (message.Message.UUID in self.regions) {
          $rootScope.$broadcast('ConsoleUpdate', message.Message);
        }
      default:
        console.log('Error parsing message:');
        console.log(message);
      }

    };*/
  };

  self.disconnect = function () {
    self.ws.io.disconnect();
    //self.ws.close();
  };

  /* location tracking */
  var locationStack = [];
  self.pushLocation = function (url) {
    locationStack.push(url);
  };
  self.popLocation = function () {
    return locationStack.pop();
  };

  /* utility functions */
  self.deleteJob = function (job) {
    return $q(function (resolve, reject) {
      self.request('DeleteJob', {
        ID: job.ID
      }, function (success, message) {
        if (success) {
          delete self.jobs[job.ID];
          resolve();
        } else {
          reject(message);
        }
      });

    });
  };
});
