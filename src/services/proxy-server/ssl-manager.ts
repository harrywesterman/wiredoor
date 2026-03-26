import path from 'path';
import FileManager from '../../utils/file-manager';
import CLI from '../../utils/cli';
import { SSLTermination, SSLCerts } from '../../database/models/domain';
import config from '../../config';
import { Logger } from '../../logger';
import { ValidationError } from '../../utils/errors/validation-error';

const selfSignedCertificatePath = '/etc/nginx/ssl';
const opensslConf = '/etc/openssl/openssl.cnf';
const cloudflareCredentialsPath = '/etc/letsencrypt/cloudflare.ini';

export class SSLManager {
  private static isWildcardDomain(domain: string): boolean {
    return !!domain && domain.startsWith('*.');
  }

  private static getCertbotDomainName(domain: string): string {
    return this.isWildcardDomain(domain) ? domain.slice(2) : domain;
  }

  private static getWildcardDnsName(domain: string): string {
    return this.isWildcardDomain(domain) ? domain : `*.${domain}`;
  }

  private static getCloudflareApiToken(): string | null {
    return config.dns.cloudflareApiToken || null;
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
    const domainCertFolder = !domain || domain === '_' ? 'default' : domain;
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

    if (
      !FileManager.isPath(`${certPath}/privkey.pem`) &&
      !FileManager.isPath(`${certPath}/fullchain.pem`)
    ) {
      const mailOption = config.admin.email
        ? `-m ${config.admin.email}`
        : '--register-unsafely-without-email';
      const isWildcard = this.isWildcardDomain(domain);

      let command = `certbot certonly --non-interactive --agree-tos --cert-name ${certName} ${mailOption}`;

      if (isWildcard) {
        await this.ensureCloudflareCredentialsFile();
        command +=
          ` --dns-cloudflare --dns-cloudflare-credentials ${cloudflareCredentialsPath}` +
          ` --dns-cloudflare-propagation-seconds 60 -d ${certName} -d ${wildcardDomain}`;
      } else {
        command += ` --webroot -w /var/www/letsencrypt -d ${domain}`;
      }

      await CLI.exec(command);
    }

    return {
      privkey: `${certPath}/privkey.pem`,
      fullchain: `${certPath}/fullchain.pem`,
    };
  }

  static async deleteCertbotCertificate(domain: string): Promise<void> {
    const certName = this.getCertbotDomainName(domain);
    const certPath = `/etc/letsencrypt/live/${certName}`;

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
      const domainCertFolder = !domain || domain === '_' ? 'default' : domain;
      return path.join(selfSignedCertificatePath, domainCertFolder);
    } else {
      return `/etc/letsencrypt/live/${this.getCertbotDomainName(domain)}`;
    }
  }
}
