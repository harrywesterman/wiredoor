export default class DomainUtils {
  static isWildcardDomain(domain: string): boolean {
    return !!domain && domain.startsWith('*.');
  }

  static getRootDomain(domain: string): string {
    if (!domain || domain === '_') {
      return 'default';
    }

    return this.isWildcardDomain(domain) ? domain.slice(2) : domain;
  }

  static getFilesystemDomainKey(domain: string): string {
    const rootDomain = this.getRootDomain(domain);

    if (rootDomain === 'default') {
      return rootDomain;
    }

    return this.isWildcardDomain(domain)
      ? `wildcard-${rootDomain}`
      : rootDomain;
  }
}
