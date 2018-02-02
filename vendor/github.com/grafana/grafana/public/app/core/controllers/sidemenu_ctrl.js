define([
  'angular',
  'lodash',
  'jquery',
  '../core_module',
  'app/core/config',
],
function (angular, _, $, coreModule, config) {
  'use strict';

  coreModule.default.controller('SideMenuCtrl', function($scope, $location, contextSrv, backendSrv) {

    $scope.getUrl = function(url) {
      return config.appSubUrl + url;
    };

    $scope.setupMainNav = function() {
      _.each(config.bootData.mainNavLinks, function(item) {
        $scope.mainLinks.push({
          text: item.text,
          icon: item.icon,
          href: $scope.getUrl(item.href)
        });
      });
    };

    $scope.loadOrgs = function() {
      $scope.orgMenu = [];

      if (contextSrv.hasRole('Admin')) {
        $scope.orgMenu.push({
          text: "Organization settings",
          href: $scope.getUrl("/org"),
        });
        $scope.orgMenu.push({
          text: "Users",
          href: $scope.getUrl("/org/users"),
        });
        $scope.orgMenu.push({
          text: "API Keys",
          href: $scope.getUrl("/org/apikeys"),
        });
      }

      if ($scope.orgMenu.length > 0) {
        $scope.orgMenu.push({ cssClass: 'divider' });
      }

      backendSrv.get('/api/user/orgs').then(function(orgs) {
        _.each(orgs, function(org) {
          if (org.orgId === contextSrv.user.orgId) {
            return;
          }

          $scope.orgMenu.push({
            text: "Switch to " + org.name,
            icon: "fa fa-fw fa-random",
            click: function() {
              $scope.switchOrg(org.orgId);
            }
          });
        });

        if (config.allowOrgCreate) {
          $scope.orgMenu.push({
            text: "New Organization",
            icon: "fa fa-fw fa-plus",
            href: $scope.getUrl('/org/new')
          });
        }
      });
    };

    $scope.switchOrg = function(orgId) {
      backendSrv.post('/api/user/using/' + orgId).then(function() {
        window.location.href = $scope.getUrl('/');
      });
    };

    $scope.setupAdminNav = function() {
      $scope.systemSection = true;
      $scope.grafanaVersion = config.buildInfo.version;

      $scope.mainLinks.push({
        text: "System info",
        icon: "fa fa-fw fa-info",
        href: $scope.getUrl("/admin/settings"),
      });

      $scope.mainLinks.push({
        text: "Global Users",
        icon: "fa fa-fw fa-user",
        href: $scope.getUrl("/admin/users"),
      });

      $scope.mainLinks.push({
        text: "Global Orgs",
        icon: "fa fa-fw fa-users",
        href: $scope.getUrl("/admin/orgs"),
      });
    };

    $scope.updateMenu = function() {
      $scope.systemSection = false;
      $scope.mainLinks = [];
      $scope.orgMenu = [];

      var currentPath = $location.path();
      if (currentPath.indexOf('/admin') === 0) {
        $scope.setupAdminNav();
      } else {
        $scope.setupMainNav();
      }
    };

    $scope.init = function() {
      $scope.showSignout = contextSrv.isSignedIn && !config['authProxyEnabled'];
      $scope.updateMenu();
      $scope.$on('$routeChangeSuccess', $scope.updateMenu);
    };
  });

});
