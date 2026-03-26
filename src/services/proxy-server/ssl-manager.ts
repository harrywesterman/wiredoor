import path from 'path';
import FileManager from '../../utils/file-manager';
import CLI from '../../utils/cli';
import { SSLTermination, SSLCerts } from '../../database/models/domain';
import config from '../../config';
import { Logger } from '../../logger';
import { ValidationError } from '../../utils/errors/validation-error';
import DomainUtils from '../../utils/domain-utils';

const selfSignedCertificatePath = '/etc/nginx/ssl';
const opensslConf = '/etc/openssl/openssl.cnf';
const cloudflareCredentialsPath = '/etc/letsencrypt/cloudflare.ini';
const certbotMetadataPath = '/etc/letsencrypt/wiredoor-metadata';

interface CertbotFamilyMetadata {
  certName: string;
  domains: string[];
}

export class SSLManager {
  private static getCertbotDomainName(domain: string): string {
    return DomainUtils.getRootDomain(domain);
  }

  private static getWildcardDnsName(domain: string): string {
    return DomainUtils.isWildcardDomain(domain) ? domain : `*.${domain}`;
  }

  private static getCloudflareApiToken(): string | null {
    return config.dns.cloudflareApiToken || null;
  }

  private static getCertbotMetadataFile(certName: string): string {
    return path.join(certbotMetadataPath, `${certName}.json`);
  }

  private static async readCertbotFamilyMetadata(
    certName: string,
  ): Promise<CertbotFamilyMetadata> {
    const metadataFile = this.getCertbotMetadataFile(certName);

    if (!FileManager.isPath(metadataFile)) {
      return {
        certName,
        domains: [],
      };
    }

    try {
      const content = await FileManager.readFile(metadataFile, 'utf-8');
      const parsed = JSON.parse(content) as CertbotFamilyMetadata;

      return {
        certName,
        domains: Array.isArray(parsed?.domains) ? parsed.domains : [],
      };
    } catch {
      return {
        certName,
        domains: [],
      };
    }
  }

  private static async writeCertbotFamilyMetadata(
    certName: string,
    domains: string[],
  ): Promise<void> {
    FileManager.mkdirSync(certbotMetadataPath);

    await FileManager.saveToFile(
      this.getCertbotMetadataFile(certName),
      `${JSON.stringify(
        {
          certName,
          domains: Array.from(new Set(domains.filter(Boolean))),
        },
        null,
        2,
      )}\n`,
      'utf-8',
      0o600,
    );
  }

  private static async addCertbotFamilyDomains(
    certName: string,
    domains: string[],
  ): Promise<void> {
    const metadata = await this.readCertbotFamilyMetadata(certName);
    const mergedDomains = Array.from(
      new Set([...metadata.domains, ...domains.filter(Boolean)]),
    );

    await this.writeCertbotFamilyMetadata(certName, mergedDomains);
  }

  private static async removeCertbotFamilyDomain(
    certName: string,
    domain: string,
  ): Promise<string[]> {
    const metadata = await this.readCertbotFamilyMetadata(certName);
    const remainingDomains = metadata.domains.filter((entry) => entry !== domain);

    if (remainingDomains.length > 0) {
      await this.writeCertbotFamilyMetadata(certName, remainingDomains);
    } else {
      await FileManager.removeFile(this.getCertbotMetadataFile(certName));
    }

    return remainingDomains;
  }

  private static async ensureCloudflareCredentialsFile(): Promise<void> {
    const apiToken = this.getCloudflareApiToken();

    if (!apiToken) {
      throw new ValidationError({
        body: [
          {
            field: 'domain',
            message:
              'Wildcard certificates require CLOUDFLARE_API_TOKEN for DNS-01 validation.',
          },
        ],
      });
    }

    await FileManager.saveToFile(
      cloudflareCredentialsPath,
      `dns_cloudflare_api_token = ${apiToken}\n`,
      'utf-8',
      0o600,
    );
  }

  static getSSLCertificates(
    domain: string,
    type: SSLTermination,
  ): Promise<SSLCerts> {
    if (!domain || domain === '_' || type === 'self-signed') {
      return this.getSelfSignedCertificates(domain);
    }

    return this.getCertbotCertificates(domain);
  }

  static async getSelfSignedCertificates(domain: string): Promise<SSLCerts> {
    const domainCertFolder = DomainUtils.getFilesystemDomainKey(domain);
    const certPath = path.join(selfSignedCertificatePath, domainCertFolder);
    if (FileManager.mkdirSync(certPath)) {
      if (
        !FileManager.isPath(`${certPath}/privkey.key`) &&
        !FileManager.isPath(`${certPath}/cert.crt`)
      ) {
        await CLI.exec(
          `openssl genpkey -algorithm RSA -out ${certPath}/privkey.key`,
        );
        await CLI.exec(
          `openssl req -new -key ${certPath}/privkey.key -out ${certPath}/wiredoor.csr -config ${opensslConf}`,
        );
        await CLI.exec(
          `openssl x509 -req -days 3650 -in ${certPath}/wiredoor.csr -signkey ${certPath}/privkey.key -out ${certPath}/cert.crt`,
        );
      }

      return {
        privkey: `${certPath}/privkey.key`,
        fullchain: `${certPath}/cert.crt`,
      };
    }
  }

  static async getCertbotCertificates(domain: string): Promise<SSLCerts> {
    const certName = this.getCertbotDomainName(domain);
    const certPath = `/etc/letsencrypt/live/${certName}`;
    const wildcardDomain = this.getWildcardDnsName(domain);
    const isWildcard = DomainUtils.isWildcardDomain(domain);
    const certExists =
      FileManager.isPath(`${certPath}/privkey.pem`) &&
      FileManager.isPath(`${certPath}/fullchain.pem`);
    const familyMetadata = await this.readCertbotFamilyMetadata(certName);
    const wildcardKnown = familyMetadata.domains.includes(wildcardDomain);

    if (!certExists || (isWildcard && !wildcardKnown)) {
      const mailOption = config.admin.email
        ? `-m ${config.admin.email}`
        : '--register-unsafely-without-email';

      let command = `certbot certonly --non-interactive --agree-tos --cert-name ${certName} ${mailOption}`;

      if (isWildcard) {
        await this.ensureCloudflareCredentialsFile();
        if (certExists) {
          command += ' --expand';
        }

        command +=
          ` --dns-cloudflare --dns-cloudflare-credentials ${cloudflareCredentialsPath}` +
          ` --dns-cloudflare-propagation-seconds 60 -d ${certName} -d ${wildcardDomain}`;
      } else {
        command += ` --webroot -w /var/www/letsencrypt -d ${domain}`;
      }

      await CLI.exec(command);
    }

    await this.addCertbotFamilyDomains(
      certName,
      isWildcard ? [certName, wildcardDomain] : [domain],
    );

    return {
      privkey: `${certPath}/privkey.pem`,
      fullchain: `${certPath}/fullchain.pem`,
    };
  }

  static async deleteCertbotCertificate(
    domain: string,
    allowDelete = true,
  ): Promise<void> {
    const certName = this.getCertbotDomainName(domain);
    const certPath = `/etc/letsencrypt/live/${certName}`;
    const remainingDomains = await this.removeCertbotFamilyDomain(
      certName,
      domain,
    );

    if (remainingDomains.length > 0 || !allowDelete) {
      return;
    }

    if (FileManager.isPath(`${certPath}/privkey.pem`)) {
      try {
        await CLI.exec(`certbot delete --cert-name ${certName} -n`);
      } catch (e: Error | any) {
        Logger.error('Certbot delete failed', e);
      }
    }
  }

  static getSSLPair(
    domain: string,
    type: SSLTermination,
  ): { privkey: string; fullchain: string } {
    const certPath = this.getCertPath(domain, type);
    if (!domain || domain === '_' || type === 'self-signed') {
      return {
        privkey: `${certPath}/privkey.key`,
        fullchain: `${certPath}/cert.crt`,
      };
    } else {
      return {
        privkey: `${certPath}/privkey.pem`,
        fullchain: `${certPath}/fullchain.pem`,
      };
    }
  }

  static getCertPath(domain: string, type: SSLTermination): string {
    if (!domain || domain === '_' || type === 'self-signed') {
      const domainCertFolder = DomainUtils.getFilesystemDomainKey(domain);
      return path.join(selfSignedCertificatePath, domainCertFolder);
    } else {
      return `/etc/letsencrypt/live/${this.getCertbotDomainName(domain)}`;
    }
  }
}
