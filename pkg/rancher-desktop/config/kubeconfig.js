'use strict';

const fs = require('fs');
const pth = require('path');

const k8s = require('@kubernetes/client-node');

// Get the path to the kubeconfig file. This is dependent on where this is run.
function path() {
  console.log(`QQQ: >> called kubeconfig.js:path()`);
  if (process.env.KUBECONFIG && process.env.KUBECONFIG.length > 0) {
    const files = process.env.KUBECONFIG.split(pth.delimiter).filter(hasAccess);

    // Only returning the path to the first file if there are multiple.
    if (files.length) {
      console.log(`QQQ: got KUBECONFIG => ${ files[0] }`);

      return files[0];
    }
  }

  const home = k8s.findHomeDir();

  console.log(`QQQ: findHomeDir => ${ home }`);

  if (home) {
    const kube = pth.join(home, '.kube');
    const cfg = pth.join(kube, 'config');

    console.log(`QQQ: k8s config from pkg/rancher-desktop/config/kubeconfig.js: ${ cfg }`);
    if (!hasAccess(cfg)) {
      if (!hasAccess(kube)) {
        console.log(`creating dir ${ kube }`);
        fs.mkdirSync(kube);
      }
      console.log(`creating file ${ cfg }`);
      fs.writeFileSync(cfg, JSON.stringify({
        apiVersion:        'v1',
        clusters:          [],
        contexts:          [],
        'current-context': null,
        kind:              'Config',
        preferences:       {},
        users:             [],
      }, undefined, 2), { mode: 0o600 });
    }

    return cfg;
  }

  // TODO: Handle WSL
  console.log(`QQQ: no KUBECONFIG, no HOME`);

  return '';
}

exports.path = path;

function hasAccess(pth) {
  try {
    fs.accessSync(pth);

    return true;
  } catch (err) {
    return false;
  }
}

exports.hasAccess = hasAccess;
