// @ts-check

import starlight from '@astrojs/starlight';
import { defineConfig } from 'astro/config';

export default defineConfig({
  integrations: [
    starlight({
      title: 'codedock-mail docs',
      sidebar: [
        {
          label: 'Start',
          items: [{ label: 'Introduction', slug: 'index' }],
        },
      ],
    }),
  ],
});
