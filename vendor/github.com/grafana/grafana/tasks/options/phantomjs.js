module.exports = function(config,grunt) {
  'use strict';

  grunt.registerTask('phantomjs', 'Copy phantomjs binary from node', function() {

    var dest = './vendor/phantomjs/phantomjs';
    var confDir = './node_modules/karma-phantomjs-launcher/node_modules/phantomjs/lib/'

    if (!grunt.file.exists(confDir)) {
      // npm 3 or npm 2 with dedupe
      confDir = './node_modules/phantomjs/lib/';
    }

    if (!grunt.file.exists(dest)){

      var m=grunt.file.read(confDir+"location.js")
      var src=/= \"([^\"]*)\"/.exec(m)[1];

      if (!grunt.file.isPathAbsolute(src)) {
        src = confDir+src;
      }

      var exec = require('child_process').execFileSync;

      try {
        grunt.config('copy.phantom_bin', {
          src: src,
          dest: dest,
          options: { mode: true},
        });
        grunt.task.run('copy:phantom_bin');
      } catch (err) {
        grunt.verbose.writeln(err);
        grunt.fail.warn('No working Phantomjs binary available')
      }

    } else {
      grunt.log.writeln('Phantomjs already imported from node');
    }
  });
};
