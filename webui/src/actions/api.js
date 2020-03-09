import base64 from 'base-64';
import {generateDownloadToken} from '../downloadToken';


const API_ENDPOINT = '/api/v1';

const basicAuth = (accessKeyId, secretAccessKey) => {
    return {
        "Authorization": `Basic ${base64.encode(`${accessKeyId}:${secretAccessKey}`)}`,
    };
};


export const linkToPath = async (repoId, branchId, path) => {
    const userData = getUser();
    const query = qs({
        path: path,
        token: await generateDownloadToken(userData.accessKeyId, userData.secretAccessKey, path),
    });
    return `${API_ENDPOINT}/repositories/${repoId}/refs/${branchId}/objects?${query}`
};

const getUser = () => {
    let userData = window.localStorage.getItem("user");
    return JSON.parse(userData);
}

const cachedBasicAuth = () => {
    const userData = getUser();
    return basicAuth(userData.accessKeyId, userData.secretAccessKey);
};

const json =(data) => {
    return JSON.stringify(data, null, "");
};

const qs = (queryParts) => {
    const parts = Object.keys(queryParts).map(key => [key, queryParts[key]]);
    return new URLSearchParams(parts).toString();
};

export const extractError = async (response) => {
    let body;
    if (response.headers.get('Content-Type') === 'application/json') {
        const jsonBody = await response.json();
        body = jsonBody.message;
    } else {
        body = await response.text();
    }
    return body;
};

const apiRequest = async (uri, requestData = {}, additionalHeaders = {}, credentials = null, defaultHeaders ={"Accept": "application/json",
    "Content-Type": "application/json",}) => {
    const auth = (credentials === null) ?
        cachedBasicAuth() : basicAuth(credentials.accessKeyId, credentials.secretAccessKey);
    return await fetch(`${API_ENDPOINT}${uri}`, {
        headers: new Headers({
            ...auth,
            ...defaultHeaders,
            ...additionalHeaders,
        }),
        ...requestData,
    });
};

// helper errors
export class NotFoundError extends Error {
    constructor(message) {
        super();
        this.message = message;
        this.name = "NotFoundError";
    }
}

// actual actions:
export const auth = {

    login: async (accessKeyId, secretAccessKey) => {
        let response = await apiRequest(  '/authentication',
            undefined,
            undefined,
            {accessKeyId, secretAccessKey});
        if (response.status === 401) {
            throw new Error('invalid credentials');
        }
        if (response.status !== 200) {
            throw new Error('unknown authentication error');
        }
        let responseJSON = await response.json();
        return {
            accessKeyId,
            secretAccessKey,
            ...responseJSON.user,
        };
    },

};

class Repositories {

    async get(repoId) {
        const response = await apiRequest(`/repositories/${repoId}`);
        if (response.status === 404) {
            throw new NotFoundError(`could not find repository ${repoId}`);
        } else if (response.status !== 200) {
            throw new Error(`could not get repository: ${await extractError(response)}`)
        }
        return await response.json();
    }

    async list(after, amount) {
        const query = qs({after, amount});
        const response = await apiRequest(`/repositories?${query}`);
        if (response.status !== 200) {
            throw new Error(`could not list repositories: ${await extractError(response)}`)
        }
        return await response.json();
    }

    async filter(from, amount) {
        if (!from) {
            return await this.list(from, amount);
        }
        const response = await this.list(from, 1000);
        let self;
        try {
            self = await this.get(from);
        } catch (error) {
            if (!(error instanceof NotFoundError)) {
                throw error;
            }
        }
        let results = response.results.filter(repo => repo.id.indexOf(from) === 0);
        let hasMore = (results.length === 1000 && response.pagination.has_more);
        if (!!self) results = [self, ...results];

        let returnVal = {
            results,
            pagination: {
                has_more: hasMore,
                max_per_page: 1000,
                results: results.length + 1,
            },
        };
        return returnVal;
    }

    async create(repo) {
        const response = await apiRequest('/repositories', {
            method: 'POST',
            body: json(repo),
        });
        if (response.status !== 201) {
            throw new Error(await extractError(response));
        }
        return await response.json();
    }

    async delete(repoId) {
        const response = await apiRequest(`/repositories/${repoId}`, {method: 'DELETE'});
        if (response.status !== 204) {
            throw new Error(await extractError(response));
        }
    }
}

class Branches {

    async get(repoId, branchId) {
        const response = await apiRequest(`/repositories/${repoId}/branches/${branchId}`);
        if (response.status === 404) {
            throw new NotFoundError(`could not find branch ${branchId}`);
        } else if (response.status !== 200) {
            throw new Error(`could not get branch: ${await extractError(response)}`)
        }
        return await response.json();
    }

    async create(repoId, branch) {
        const response = await apiRequest(`/repositories/${repoId}/branches`, {
            method: 'POST',
            body: json(branch),
        });
        if (response.status !== 201) {
            throw new Error(await extractError(response));
        }
        return await response.json();
    }

    async list(repoId, after, amount) {
        const query = qs({after, amount});
        const response = await apiRequest(`/repositories/${repoId}/branches?${query}`);
        if (response.status !== 200) {
            throw new Error(`could not list branches: ${await extractError(response)}`)
        }
        return await response.json();
    }
    async filter(repoId, from, amount) {
        if (!from) {
            return await this.list(repoId, from, amount);
        }
        const response = await this.list(repoId, from, amount);
        let self;
        try {
            self = await this.get(repoId, from);
        } catch (error) {
            if (!(error instanceof NotFoundError)) {
                throw error;
            }
        }
        let results = response.results.filter(branch => branch.id.indexOf(from) === 0);
        let hasMore = (results.length === amount && response.pagination.has_more);
        if (!!self) results = [self, ...results];

        let returnVal = {
            results,
            pagination: {
                has_more: hasMore,
                max_per_page: 1000,
                results: results.length + (!!self) ? 1 : 0,
            },
        };
        return returnVal;
    }
}

class Objects {

    async list(repoId, ref, tree, after= "", amount = 1000) {
        const query = qs({tree, amount, after});
        const response = await apiRequest(`/repositories/${repoId}/refs/${ref}/objects/ls?${query}`);
        if (response.status !== 200) {
            throw new Error(await extractError(response));
        }
        return await response.json();
    }

    async upload(repoId, branchId, path, fileObject) {
        const data = new FormData();
        data.append('content', fileObject);
        const query = qs({path});
        const response = await apiRequest(`/repositories/${repoId}/branches/${branchId}/objects?${query}`, {
            method: 'POST',
            body: data,
        }, {}, null, {});
        if (response.status !== 201) {
            throw new Error(await extractError(response));
        }
        return await response.json();
    }
}

class Commits {
    async log(repoId, branchId) {
        const response = await apiRequest(`/repositories/${repoId}/branches/${branchId}/commits`);
        if (response.status !== 200) {
            throw new Error(await extractError(response));
        }
        const data = await response.json();
        return data.results;
    }
}

class Refs {
    async diff(repoId, leftRef, rightRef) {
        let response;
        if (leftRef === rightRef) {
            response = await apiRequest(`/repositories/${repoId}/branches/${leftRef}/diff`);
        } else {
            response = await apiRequest(`/repositories/${repoId}/refs/${leftRef}/diff/${rightRef}`);
        }
        if (response.status !== 200) {
            throw new Error(await extractError(response));
        }
        return await response.json();
    }
}

export const repositories = new Repositories();
export const branches = new Branches();
export const objects = new Objects();
export const commits = new Commits();
export const refs = new Refs();