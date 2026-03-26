import Container from 'typedi';
import { loadApp } from '../../main';
import { DataSource } from 'typeorm';
import {
  mockCLIExec,
  mockIsPath,
  mockNslookup,
  mockRemoveDir,
  mockRemoveFile,
  mockSaveToFile,
} from '../.jest/global-mocks';
import config from '../../config';
import { DomainRepository } from '../../repositories/domain-repository';
import { DomainsService } from '../../services/domains-service';
import { DomainQueryFilter } from '../../repositories/filters/domain-query-filter';
import { makeDomainData } from './stubs/domain.stub';
import { PagedData } from '../../repositories/filters/repository-query-filter';
import { Domain } from '../../database/models/domain';
import { SSLManager } from '../../services/proxy-server/ssl-manager';
import DomainUtils from '../../utils/domain-utils';
import FileManager from '../../utils/file-manager';

// eslint-disable-next-line @typescript-eslint/no-unused-vars
let app;
let dataSource: DataSource;

beforeAll(async () => {
  app = await loadApp();
  dataSource = Container.get('dataSource');
});

afterAll(async () => {});

describe('Domains Service', () => {
  let repository: DomainRepository;
  let service: DomainsService;
  let filter: DomainQueryFilter;

  beforeEach(async () => {
    repository = new DomainRepository(dataSource);
    filter = new DomainQueryFilter(repository);
    service = new DomainsService(repository, filter);

    jest.clearAllMocks();
  });

  afterEach(async () => {
    await repository.clear();
    jest.clearAllMocks();
  });

  describe('Initialization', () => {
    it('should update server configuration files for all Domains when initialize', async () => {
      const domainData1 = makeDomainData({ ssl: 'self-signed' });
      const domainData2 = makeDomainData({ ssl: 'certbot' });

      await repository.save(domainData1);
      await repository.save(domainData2);

      mockNslookup.mockImplementation(
        jest.fn(() => {
          return true;
        }),
      );

      jest.clearAllMocks();

      await service.initialize();

      expect(mockSaveToFile.mock.calls).toEqual([
        [
          `/etc/nginx/conf.d/${domainData1.domain}.conf`,
          expect.stringContaining(` ${domainData1.domain};`),
        ],
        [
          `/etc/nginx/locations/${domainData1.domain}/__main.conf`,
          expect.stringContaining(`root /etc/nginx/default_pages;`),
        ],
        [
          `/etc/nginx/conf.d/${domainData2.domain}.conf`,
          expect.stringContaining(` ${domainData2.domain};`),
        ],
        [
          `/etc/nginx/locations/${domainData2.domain}/__main.conf`,
          expect.stringContaining(`root /etc/nginx/default_pages;`),
        ],
      ]);

      expect(mockCLIExec.mock.calls).toEqual([
        [
          expect.stringMatching(
            new RegExp(`openssl.*${domainData1.domain?.replace('.', '\\.')}.*`),
          ),
        ],
        [
          expect.stringMatching(
            new RegExp(`openssl.*${domainData1.domain?.replace('.', '\\.')}.*`),
          ),
        ],
        [
          expect.stringMatching(
            new RegExp(`openssl.*${domainData1.domain?.replace('.', '\\.')}.*`),
          ),
        ],
        ['nginx -t'],
        [
          expect.stringMatching(
            new RegExp(`certbot.*${domainData2.domain?.replace('.', '\\.')}.*`),
          ),
        ],
        ['nginx -t'],
        ['supervisorctl reread && supervisorctl update'],
      ]);
    });
  });

  describe('List Domains', () => {
    it('should list all Domains', async () => {
      const data = makeDomainData();
      const domain = await repository.save({ ...data });

      const result = await service.getDomains({});

      expect(result).toEqual(
        expect.arrayContaining([
          expect.objectContaining({ domain: data.domain }),
        ]),
      );
    });
    it('should list Domains paginated', async () => {
      const data = makeDomainData();
      await repository.save({ ...data });

      const result = await service.getDomains({ limit: 1 });

      expect((result as PagedData<Domain>).data.length).toEqual(1);
    });
  });

  describe('Create Domain', () => {
    it('should create Domain and save server configuration file', async () => {
      const data = makeDomainData({ ssl: 'self-signed' });

      mockNslookup.mockImplementation(
        jest.fn(() => {
          return false;
        }),
      );

      jest.clearAllMocks();

      const result = await service.createDomain(data);

      expect(result.id).toBeDefined();
      expect(result.domain).toEqual(data.domain);

      expect(mockSaveToFile.mock.calls).toEqual([
        [
          `/etc/nginx/conf.d/${data.domain}.conf`,
          expect.stringContaining(` ${data.domain};`),
        ],
        [
          `/etc/nginx/locations/${data.domain}/__main.conf`,
          expect.stringContaining(`root /etc/nginx/default_pages;`),
        ],
      ]);

      expect(mockCLIExec.mock.calls).toEqual([
        [
          expect.stringMatching(
            new RegExp(`openssl.*${data.domain?.replace('.', '\\.')}.*`),
          ),
        ],
        [
          expect.stringMatching(
            new RegExp(`openssl.*${data.domain?.replace('.', '\\.')}.*`),
          ),
        ],
        [
          expect.stringMatching(
            new RegExp(`openssl.*${data.domain?.replace('.', '\\.')}.*`),
          ),
        ],
        ['nginx -t'],
        ['nginx -s reload'],
      ]);
    });
    it('should create Domain and call certbot', async () => {
      const data = makeDomainData({ ssl: 'certbot' });

      mockNslookup.mockImplementation(
        jest.fn(() => {
          return true;
        }),
      );

      jest.clearAllMocks();

      const result = await service.createDomain(data);

      expect(result.id).toBeDefined();
      expect(result.domain).toEqual(data.domain);

      expect(mockSaveToFile.mock.calls).toEqual([
        [
          `/etc/nginx/conf.d/${data.domain}.conf`,
          expect.stringContaining(` ${data.domain};`),
        ],
        [
          `/etc/nginx/locations/${data.domain}/__main.conf`,
          expect.stringContaining(`root /etc/nginx/default_pages;`),
        ],
      ]);

      expect(mockCLIExec.mock.calls).toEqual([
        [
          expect.stringMatching(
            new RegExp(`certbot.*${data.domain?.replace('.', '\\.')}.*`),
          ),
        ],
        ['nginx -t'],
        ['nginx -s reload'],
      ]);
    });
    it('should create domain with authentication enabled', async () => {
      const data = makeDomainData({ ssl: 'certbot', authentication: true });

      jest.clearAllMocks();

      const result = await service.createDomain(data);

      expect(result.oauth2ServicePort).toEqual(expect.any(Number));
      expect(result.id).toBeDefined();
      expect(result.domain).toEqual(data.domain);

      expect(mockSaveToFile.mock.calls).toEqual([
        [
          `/data/oauth2/.cookie-secret-${data.domain}`,
          expect.any(String),
          'utf-8',
          0o600,
        ],
        [
          `/opt/oauth2-proxy/${data.domain}-emails`,
          result.oauth2Config.allowedEmails.join('\n'),
          'utf-8',
          0o644,
        ],
        [
          `/etc/supervisor/conf.d/oauth2-proxy-d${result.id}.conf`,
          expect.stringContaining(
            ` OAUTH2_PROXY_HTTP_ADDRESS="127.0.0.1:${result.oauth2ServicePort}",`,
          ),
          'utf-8',
          0o644,
        ],
        [
          `/etc/nginx/conf.d/${data.domain}.conf`,
          expect.stringContaining(
            ` http://127.0.0.1:${result.oauth2ServicePort};`,
          ),
        ],
        [
          `/etc/nginx/locations/${data.domain}/__main.conf`,
          expect.stringContaining(`root /etc/nginx/default_pages;`),
        ],
      ]);

      expect(mockCLIExec.mock.calls).toEqual([
        [
          expect.stringMatching(
            new RegExp(`certbot.*${data.domain?.replace('.', '\\.')}.*`),
          ),
        ],
        ['supervisorctl reread && supervisorctl update'],
        ['nginx -t'],
        ['nginx -s reload'],
      ]);
    });

    it('should create a wildcard domain using Cloudflare DNS-01 and verify a concrete subdomain', async () => {
      const data = makeDomainData({
        domain: '*.example.com',
        ssl: 'certbot',
      });

      const originalProvider = config.dns.provider;
      const originalToken = config.dns.cloudflareApiToken;
      const dnsService = {
        canManageDomain: jest.fn().mockResolvedValue(true),
        createRecord: jest.fn().mockResolvedValue({
          type: 'A',
          name: '*.example.com',
          content: '203.0.113.10',
          ttl: 3600,
        }),
        waitUntilResolvesTo: jest.fn().mockResolvedValue(undefined),
      };

      config.dns.provider = 'cloudflare';
      config.dns.cloudflareApiToken = 'test-cloudflare-token';
      mockNslookup.mockImplementation(
        jest.fn(() => {
          return false;
        }),
      );
      service['dnsService'] = dnsService as any;

      try {
        const result = await service.createDomainIfNotExists(data.domain);
        const filesystemDomainKey = DomainUtils.getFilesystemDomainKey(
          data.domain,
        );

        expect(result.domain).toEqual(data.domain);
        expect(result.ssl).toEqual('certbot');
        expect(result.sslPair.fullchain).toEqual(
          '/etc/letsencrypt/live/example.com/fullchain.pem',
        );
        expect(mockSaveToFile).toHaveBeenCalledWith(
          `/etc/nginx/conf.d/${filesystemDomainKey}.conf`,
          expect.stringContaining(` ${data.domain};`),
        );
        expect(mockSaveToFile).toHaveBeenCalledWith(
          `/etc/nginx/locations/${filesystemDomainKey}/__main.conf`,
          expect.stringContaining(`root /etc/nginx/default_pages;`),
        );
        expect(dnsService.createRecord).toHaveBeenCalledWith({
          name: '*.example.com',
          type: 'A',
          content: '203.0.113.10',
          ttl: 3600,
          proxied: false,
        });
        expect(dnsService.waitUntilResolvesTo).toHaveBeenCalledWith(
          'wiredoor-verify.example.com',
          '203.0.113.10',
          expect.objectContaining({
            timeoutMs: 30_000,
            intervalMs: 1_000,
          }),
        );
        expect(mockSaveToFile).toHaveBeenCalledWith(
          '/etc/letsencrypt/cloudflare.ini',
          expect.stringContaining(
            'dns_cloudflare_api_token = test-cloudflare-token',
          ),
          'utf-8',
          0o600,
        );
        expect(mockCLIExec.mock.calls[0][0]).toContain('--dns-cloudflare');
        expect(mockCLIExec.mock.calls[0][0]).toContain(
          '--dns-cloudflare-credentials /etc/letsencrypt/cloudflare.ini',
        );
        expect(mockCLIExec.mock.calls[0][0]).toContain(
          '-d example.com -d *.example.com',
        );
      } finally {
        config.dns.provider = originalProvider;
        config.dns.cloudflareApiToken = originalToken;
      }
    });

    it('should reject wildcard certificates when Cloudflare DNS-01 is missing', async () => {
      await expect(
        SSLManager.getSSLCertificates('*.example.com', 'certbot' as any),
      ).rejects.toMatchObject({
        errors: {
          body: [
            expect.objectContaining({
              field: 'domain',
              message: expect.stringContaining('CLOUDFLARE_API_TOKEN'),
            }),
          ],
        },
      });
    });

    it('should expand an existing apex lineage to include the wildcard SAN', async () => {
      const readFileSpy = jest
        .spyOn(FileManager, 'readFile')
        .mockResolvedValue(
          JSON.stringify({
            certName: 'example.com',
            domains: ['example.com'],
          }),
        );

      (mockIsPath as unknown as jest.Mock).mockImplementation(
        (target: string) => {
          return (
            target.startsWith('/etc/letsencrypt/live/example.com/') ||
            target === '/etc/letsencrypt/wiredoor-metadata/example.com.json'
          );
        },
      );

      const originalProvider = config.dns.provider;
      const originalToken = config.dns.cloudflareApiToken;
      config.dns.provider = 'cloudflare';
      config.dns.cloudflareApiToken = 'test-cloudflare-token';

      try {
        await SSLManager.getSSLCertificates('*.example.com', 'certbot' as any);

        expect(mockCLIExec.mock.calls[0][0]).toContain('--expand');
        expect(mockCLIExec.mock.calls[0][0]).toContain(
          '-d example.com -d *.example.com',
        );
      } finally {
        readFileSpy.mockRestore();
        config.dns.provider = originalProvider;
        config.dns.cloudflareApiToken = originalToken;
      }
    });
  });

  describe('Update Domain', () => {
    it('should update Domain and save server configuration file', async () => {
      const data = makeDomainData();

      const created = await service.createDomain(data);

      jest.clearAllMocks();

      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const result = await service.updateDomain(created.id, {
        ssl: 'certbot',
        domain: created.domain,
      });

      // expect(result.domain).toEqual(update.domain);

      expect(mockSaveToFile.mock.calls).toEqual([
        [
          `/etc/nginx/conf.d/${data.domain}.conf`,
          expect.stringContaining(` ${data.domain};`),
        ],
        [
          `/etc/nginx/locations/${data.domain}/__main.conf`,
          expect.stringContaining(`root /etc/nginx/default_pages;`),
        ],
      ]);

      expect(mockCLIExec.mock.calls).toEqual([
        [
          expect.stringMatching(
            new RegExp(`certbot.*${data.domain?.replace('.', '\\.')}.*`),
          ),
        ],
        ['nginx -t'],
        ['nginx -s reload'],
      ]);
    });
  });

  describe('Delete Domain', () => {
    it('should delete Domain and server config file', async () => {
      const data = makeDomainData();

      mockIsPath.mockImplementation(() => {
        return true;
      });
      const created = await service.createDomain(data);

      jest.clearAllMocks();

      await service.deleteDomain(created.id);

      expect(mockRemoveFile).toHaveBeenCalledWith(
        expect.stringContaining(`/${data.domain}.conf`),
      );

      if (data.ssl === 'certbot') {
        expect(mockCLIExec).toHaveBeenCalledWith(
          `certbot delete --cert-name ${data.domain} -n`,
        );
      } else {
        expect(mockRemoveDir).toHaveBeenCalledWith(
          expect.stringContaining(`/${data.domain}`),
        );
      }

      expect(mockCLIExec).toHaveBeenCalledWith('nginx -s reload');
    });

    it('should keep the shared certbot lineage alive until the last apex or wildcard owner is removed', async () => {
      const originalProvider = config.dns.provider;
      const originalToken = config.dns.cloudflareApiToken;
      config.dns.provider = 'cloudflare';
      config.dns.cloudflareApiToken = 'test-cloudflare-token';

      const wildcardDomain = makeDomainData({
        domain: '*.example.com',
        ssl: 'certbot',
      });
      const apexDomain = makeDomainData({
        domain: 'example.com',
        ssl: 'certbot',
      });

      const wildcardCreated = await service.createDomain(wildcardDomain);
      const apexCreated = await service.createDomain(apexDomain);

      jest.clearAllMocks();
      (mockIsPath as unknown as jest.Mock).mockImplementation(
        (target: string) => {
          return target.startsWith('/etc/letsencrypt/live/example.com/');
        },
      );

      try {
        await service.deleteDomain(
          wildcardCreated.id,
        );

        expect(mockRemoveFile).toHaveBeenCalledWith(
          '/etc/nginx/conf.d/wildcard-example.com.conf',
        );
        expect(mockCLIExec.mock.calls.some(([cmd]) =>
          `${cmd}`.includes('certbot delete'),
        )).toBe(false);

        jest.clearAllMocks();

        await service.deleteDomain(apexCreated.id);

        expect(mockRemoveFile).toHaveBeenCalledWith(
          '/etc/nginx/conf.d/example.com.conf',
        );
        expect(mockCLIExec).toHaveBeenCalledWith(
          'certbot delete --cert-name example.com -n',
        );
      } finally {
        config.dns.provider = originalProvider;
        config.dns.cloudflareApiToken = originalToken;
      }
    });
  });
});
