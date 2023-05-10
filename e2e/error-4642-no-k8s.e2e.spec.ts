import path from 'path';

import { test, expect, _electron } from '@playwright/test';

import { NavPage } from './pages/nav-page';
import {
  createDefaultSettings, reportAsset, retry, teardown, tool,
} from './utils/TestUtils';

import { ContainerEngine } from '@pkg/config/settings';

import type { ElectronApplication, BrowserContext, Page } from '@playwright/test';

test.describe.serial('KubernetesBackend', () => {
  let electronApp: ElectronApplication;
  let context: BrowserContext;
  let page: Page;

  // This still needs a factory-reset...
  test.beforeAll(async() => {
    createDefaultSettings(
      {
        kubernetes:      { enabled: false },
        containerEngine: {
          name:          ContainerEngine.CONTAINERD,
          allowedImages: { enabled: false },
        },
        application: { adminAccess: false },
      });

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

    await context.tracing.start({ screenshots: true, snapshots: true });
    page = await electronApp.firstWindow();
  });

  test.afterAll(() => teardown(electronApp, __filename));

  test('should start loading the background services and hide progress bar', async() => {
    const navPage = new NavPage(page);

    await navPage.progressBecomesReady();
    await expect(navPage.progressBar).toBeHidden();
  });

  test.describe('containerd startup', () => {
    test('should emit connection information', async() => {
      await retry(async() => {
        const info = await tool('nerdctl', 'info');

        expect(info).toContain('Server Version:');
      });
    });
    test('should have buildkitd status started', async() => {
      let i = 0;

      await retry(async() => {
        const info = await tool('rdctl', 'shell', 'rc-service', '--nocolor', 'buildkitd', 'status');

        if (info.includes('unsupervised')) {
          i += 1;
          console.log(`Found unsupervised buildkitd status: ${ i }`);
        }
        expect(info).toContain('status: started');
      });
    });
  });
});
