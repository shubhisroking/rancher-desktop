// Test case 30

import fs from 'fs';
import path from 'path';

import { test, expect, _electron } from '@playwright/test';

import { NavPage } from './pages/nav-page';
import {
  createDefaultSettings, reportAsset, retry, teardown, tool,
} from './utils/TestUtils';

import { spawnFile } from '@pkg/utils/childProcess';

import type { ElectronApplication, BrowserContext, Page } from '@playwright/test';

let originalPathManagement = '';
let bashrc = '';
const adjustedPath = (process.env.PATH ?? '').split(path.delimiter).filter(dir => !dir.includes('/.rd/bin')).join(path.delimiter);
const adjustedEnv = {
  ...process.env,
  PATH: adjustedPath,
};

test.describe.serial('KubernetesBackend', () => {
  let electronApp: ElectronApplication;
  let context: BrowserContext;
  let page: Page;
  const commands: Record<string, string[]> = {
    bash: ['-l', '-c', 'which rdctl'],
    ksh:  ['-c', 'which rdctl'],
    zsh:  ['-i', '-c', 'which rdctl'],
    fish: ['-c', 'which rdctl'],
  };

  test.beforeAll(async() => {
    createDefaultSettings();
    const homedir = process.env['HOME'] ?? '.';

    bashrc = path.join(homedir, '.bashrc');
    try {
      await fs.promises.access(bashrc, fs.constants.R_OK);
    } catch {
      console.log(`Can't find a ~/.bashrc file`);
      bashrc = '';
      process.exit(2);
    }

    electronApp = await _electron.launch({
      args: [
        path.join(__dirname, '../'),
        '--disable-gpu',
        '--whitelisted-ips=',
        // See pkg/rancher-desktop/utils/commandLine.ts before changing the next item as the final option.
        '--disable-dev-shm-usage',
        '--no-modal-dialogs',
      ],
      env: {
        ...process.env,
        RD_LOGS_DIR: reportAsset(__filename, 'log'),
      },
    });
    context = electronApp.context();

    await context.tracing.start({
      screenshots: true,
      snapshots:   true,
    });
    page = await electronApp.firstWindow();
    const currentPath = process.env['PATH'] ?? '';

    process.env['PATH'] = currentPath.split(path.sep)
      .filter(x => !x.includes('.rd/bin'))
      .join(path.sep);
    originalPathManagement = JSON.parse(await tool('rdctl', 'list-settings')).application.pathManagement;
  });

  test.afterAll(() => {
    if (bashrc !== '') {
      teardown(electronApp, __filename);
    }
  });

  test('should start loading the background services and hide progress bar', async() => {
    const navPage = new NavPage(page);

    await navPage.progressBecomesReady();
    await expect(navPage.progressBar).toBeHidden();
  });

  test('ensure path-management is managed', async() => {
    if (originalPathManagement !== 'rcfiles') {
      await tool('rdctl', 'set', '--application.path-management-strategy=rcfiles');
    }
  });

  test('find out which shells are supported', async() => {
    for (const shellName in commands) {
      const args = commands[shellName];

      try {
        const result = await spawnFile(shellName, args, { env: adjustedEnv, stdio: 'pipe' });

        if (result.stderr) {
          delete commands[shellName];
        } else if (!result.stdout.includes('.rd/bin/rdctl')) {
          delete commands[shellName];
        }
      } catch {
        delete commands[shellName];
      }
    }
    expect(Object.keys(commands).length).toBeGreaterThan(0);
  });

  test('can switch to manual path-management', async() => {
    if (originalPathManagement !== 'rcfiles') {
      await tool('rdctl', 'set', '--application.path-management-strategy=manual');
      await retry(() => spawnFile('grep', ['--files-without-match', 'PATH=.*\\.rd/bin', bashrc]), { delay: 10, tries: 10 });
    }
  });

  test('no supported shells work in manual mode', async() => {
    for (const shellName in commands) {
      const args = commands[shellName];

      try {
        const result = await spawnFile(shellName, args, { env: adjustedEnv, stdio: 'pipe' });

        expect(result).toMatchObject({
          stdout: '',
          stderr: expect.stringMatching(/\S/),
        });
      } catch { }
    }
  });

  test('can switch back to rcfile path-management', async() => {
    await tool('rdctl', 'set', '--application.path-management-strategy=rcfiles');
    await retry(() => spawnFile('grep', ['--files-with-matches', '--no-messages', 'PATH=.*\\.rd/bin', bashrc]), {
      delay: 10,
      tries: 10,
    });
  });

  test('the supported shells work in rcfiles mode', async() => {
    for (const shellName in commands) {
      const args = commands[shellName];

      await expect(spawnFile(shellName, args, { env: adjustedEnv, stdio: 'pipe' })).resolves
        .toMatchObject({
          stdout: expect.stringContaining('/.rd/bin'),
          stderr: '',
        });
    }
  });
});
