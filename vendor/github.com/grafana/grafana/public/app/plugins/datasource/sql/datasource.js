define([
  'angular',
],
function (angular) {
  'use strict';

  var module = angular.module('grafana.services');

  module.factory('SqlDatasource', function() {

    function SqlDatasource() {
    }

    return SqlDatasource;

  });

});
