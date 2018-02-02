define([
  'angular',
  'jquery',
  'lodash',
  'moment',
],
function (angular, $, _, moment) {
  'use strict';

  var module = angular.module('grafana.services');

  module.factory('dashboardSrv', function()  {

    function DashboardModel (data, meta) {
      if (!data) {
        data = {};
      }

      if (!data.id && data.version) {
        data.schemaVersion = data.version;
      }

      this.id = data.id || null;
      this.title = data.title || 'No Title';
      this.originalTitle = this.title;
      this.tags = data.tags || [];
      this.style = data.style || "dark";
      this.timezone = data.timezone || 'browser';
      this.editable = data.editable !== false;
      this.hideControls = data.hideControls || false;
      this.sharedCrosshair = data.sharedCrosshair || false;
      this.rows = data.rows || [];
      this.time = data.time || { from: 'now-6h', to: 'now' };
      this.timepicker = data.timepicker || {};
      this.templating = this._ensureListExist(data.templating);
      this.annotations = this._ensureListExist(data.annotations);
      this.refresh = data.refresh;
      this.snapshot = data.snapshot;
      this.schemaVersion = data.schemaVersion || 0;
      this.version = data.version || 0;
      this.links = data.links || [];
      this._updateSchema(data);
      this._initMeta(meta);
    }

    var p = DashboardModel.prototype;

    p._initMeta = function(meta) {
      meta = meta || {};

      meta.canShare = meta.canShare !== false;
      meta.canSave = meta.canSave !== false;
      meta.canStar = meta.canStar !== false;
      meta.canEdit = meta.canEdit !== false;

      if (!this.editable) {
        meta.canEdit = false;
        meta.canDelete = false;
        meta.canSave = false;
        this.hideControls = true;
      }

      this.meta = meta;
    };

    // cleans meta data and other non peristent state
    p.getSaveModelClone = function() {
      var copy = angular.copy(this);
      delete copy.meta;
      return copy;
    };

    p._ensureListExist = function (data) {
      if (!data) { data = {}; }
      if (!data.list) { data.list = []; }
      return data;
    };

    p.getNextPanelId = function() {
      var i, j, row, panel, max = 0;
      for (i = 0; i < this.rows.length; i++) {
        row = this.rows[i];
        for (j = 0; j < row.panels.length; j++) {
          panel = row.panels[j];
          if (panel.id > max) { max = panel.id; }
        }
      }
      return max + 1;
    };

    p.forEachPanel = function(callback) {
      var i, j, row;
      for (i = 0; i < this.rows.length; i++) {
        row = this.rows[i];
        for (j = 0; j < row.panels.length; j++) {
          callback(row.panels[j], j, row, i);
        }
      }
    };

    p.getPanelById = function(id) {
      for (var i = 0; i < this.rows.length; i++) {
        var row = this.rows[i];
        for (var j = 0; j < row.panels.length; j++) {
          var panel = row.panels[j];
          if (panel.id === id) {
            return panel;
          }
        }
      }
      return null;
    };

    p.rowSpan = function(row) {
      return _.reduce(row.panels, function(p,v) {
        return p + v.span;
      },0);
    };

    p.addPanel = function(panel, row) {
      var rowSpan = this.rowSpan(row);
      var panelCount = row.panels.length;
      var space = (12 - rowSpan) - panel.span;
      panel.id = this.getNextPanelId();

      // try to make room of there is no space left
      if (space <= 0) {
        if (panelCount === 1) {
          row.panels[0].span = 6;
          panel.span = 6;
        }
        else if (panelCount === 2) {
          row.panels[0].span = 4;
          row.panels[1].span = 4;
          panel.span = 4;
        }
      }

      row.panels.push(panel);
    };

    p.isSubmenuFeaturesEnabled = function() {
      return this.templating.list.length > 0 || this.annotations.list.length > 0 || this.links.length > 0;
    };

    p.getPanelInfoById = function(panelId) {
      var result = {};
      _.each(this.rows, function(row) {
        _.each(row.panels, function(panel, index) {
          if (panel.id === panelId) {
            result.panel = panel;
            result.row = row;
            result.index = index;
          }
        });
      });

      if (!result.panel) {
        return null;
      }

      return result;
    };

    p.duplicatePanel = function(panel, row) {
      var rowIndex = _.indexOf(this.rows, row);
      var newPanel = angular.copy(panel);
      newPanel.id = this.getNextPanelId();

      delete newPanel.repeat;
      delete newPanel.repeatIteration;
      delete newPanel.repeatPanelId;
      delete newPanel.scopedVars;

      var currentRow = this.rows[rowIndex];
      currentRow.panels.push(newPanel);
      return newPanel;
    };

    p.getNextQueryLetter = function(panel) {
      var letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';

      return _.find(letters, function(refId) {
        return _.every(panel.targets, function(other) {
          return other.refId !== refId;
        });
      });
    };

    p.addDataQueryTo = function(panel, datasource) {
      var target = {
        refId: this.getNextQueryLetter(panel)
      };

      if (datasource) {
        target.datasource = datasource.name;
      }

      panel.targets.push(target);
    };

    p.removeDataQuery = function (panel, query) {
      panel.targets = _.without(panel.targets, query);
    };

    p.duplicateDataQuery = function(panel, query) {
      var clone = angular.copy(query);
      clone.refId = this.getNextQueryLetter(panel);
      panel.targets.push(clone);
    };

    p.moveDataQuery = function(panel, fromIndex, toIndex) {
      _.move(panel.targets, fromIndex, toIndex);
    };

    p.formatDate = function(date, format) {
      date = moment.isMoment(date) ? date : moment(date);
      format = format || 'YYYY-MM-DD HH:mm:ss';

      return this.timezone === 'browser' ?
        moment(date).format(format) :
        moment.utc(date).format(format);
    };

    p.getRelativeTime = function(date) {
      date = moment.isMoment(date) ? date : moment(date);

      return this.timezone === 'browser' ?
        moment(date).fromNow() :
        moment.utc(date).fromNow();
    };

    p._updateSchema = function(old) {
      var i, j, k;
      var oldVersion = this.schemaVersion;
      var panelUpgrades = [];
      this.schemaVersion = 8;

      if (oldVersion === 8) {
        return;
      }

      // version 2 schema changes
      if (oldVersion < 2) {

        if (old.services) {
          if (old.services.filter) {
            this.time = old.services.filter.time;
            this.templating.list = old.services.filter.list || [];
          }
          delete this.services;
        }

        panelUpgrades.push(function(panel) {
          // rename panel type
          if (panel.type === 'graphite') {
            panel.type = 'graph';
          }

          if (panel.type !== 'graph') {
            return;
          }

          if (_.isBoolean(panel.legend)) { panel.legend = { show: panel.legend }; }

          if (panel.grid) {
            if (panel.grid.min) {
              panel.grid.leftMin = panel.grid.min;
              delete panel.grid.min;
            }

            if (panel.grid.max) {
              panel.grid.leftMax = panel.grid.max;
              delete panel.grid.max;
            }
          }

          if (panel.y_format) {
            panel.y_formats[0] = panel.y_format;
            delete panel.y_format;
          }

          if (panel.y2_format) {
            panel.y_formats[1] = panel.y2_format;
            delete panel.y2_format;
          }
        });
      }

      // schema version 3 changes
      if (oldVersion < 3) {
        // ensure panel ids
        var maxId = this.getNextPanelId();
        panelUpgrades.push(function(panel) {
          if (!panel.id) {
            panel.id = maxId;
            maxId += 1;
          }
        });
      }

      // schema version 4 changes
      if (oldVersion < 4) {
        // move aliasYAxis changes
        panelUpgrades.push(function(panel) {
          if (panel.type !== 'graph') { return; }
          _.each(panel.aliasYAxis, function(value, key) {
            panel.seriesOverrides = [{ alias: key, yaxis: value }];
          });
          delete panel.aliasYAxis;
        });
      }

      if (oldVersion < 6) {
        // move pulldowns to new schema
        var annotations = _.findWhere(old.pulldowns, { type: 'annotations' });

        if (annotations) {
          this.annotations = {
            list: annotations.annotations || [],
          };
        }

        // update template variables
        for (i = 0 ; i < this.templating.list.length; i++) {
          var variable = this.templating.list[i];
          if (variable.datasource === void 0) { variable.datasource = null; }
          if (variable.type === 'filter') { variable.type = 'query'; }
          if (variable.type === void 0) { variable.type = 'query'; }
          if (variable.allFormat === void 0) { variable.allFormat = 'glob'; }
        }
      }

      if (oldVersion < 7) {
        if (old.nav && old.nav.length) {
          this.timepicker = old.nav[0];
          delete this.nav;
        }

        // ensure query refIds
        panelUpgrades.push(function(panel) {
          _.each(panel.targets, function(target) {
            if (!target.refId) {
              target.refId = this.getNextQueryLetter(panel);
            }
          }, this);
        });
      }

      if (oldVersion < 8) {
        panelUpgrades.push(function(panel) {
          _.each(panel.targets, function(target) {
            // update old influxdb query schema
            if (target.fields && target.tags && target.groupBy) {
              if (target.rawQuery) {
                delete target.fields;
                delete target.fill;
              } else {
                target.select = _.map(target.fields, function(field) {
                  var parts = [];
                  parts.push({type: 'field', params: [field.name]});
                  parts.push({type: field.func, params: []});
                  if (field.mathExpr) {
                    parts.push({type: 'math', params: [field.mathExpr]});
                  }
                  if (field.asExpr) {
                    parts.push({type: 'alias', params: [field.asExpr]});
                  }
                  return parts;
                });
                delete target.fields;
                _.each(target.groupBy, function(part) {
                  if (part.type === 'time' && part.interval)  {
                    part.params = [part.interval];
                    delete part.interval;
                  }
                  if (part.type === 'tag' && part.key) {
                    part.params = [part.key];
                    delete part.key;
                  }
                });

                if (target.fill) {
                  target.groupBy.push({type: 'fill', params: [target.fill]});
                  delete target.fill;
                }
              }
            }
          });
        });
      }

      if (panelUpgrades.length === 0) {
        return;
      }

      for (i = 0; i < this.rows.length; i++) {
        var row = this.rows[i];
        for (j = 0; j < row.panels.length; j++) {
          for (k = 0; k < panelUpgrades.length; k++) {
            panelUpgrades[k].call(this, row.panels[j]);
          }
        }
      }
    };

    return {
      create: function(dashboard, meta) {
        return new DashboardModel(dashboard, meta);
      },
      setCurrent: function(dashboard) {
        this.currentDashboard = dashboard;
      },
      getCurrent: function() {
        return this.currentDashboard;
      },
    };
  });
});
