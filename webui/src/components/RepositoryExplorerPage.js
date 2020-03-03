import React, {useEffect} from "react";
import {useParams, useLocation, Switch, Route, useRouteMatch, Link, generatePath, Redirect} from "react-router-dom";

import Breadcrumb from "react-bootstrap/Breadcrumb";
import Octicon, {GitCommit, GitBranch, Gear, Database}  from "@primer/octicons-react";

import TreePage from './TreePage';
import CommitsPage from './CommitsPage';
import {connect} from "react-redux";
import {getRepository} from "../actions/repositories";
import Nav from "react-bootstrap/Nav";
import Alert from "react-bootstrap/Alert";


function useQuery() {
    return new URLSearchParams(useLocation().search);
}

const qs = (queryParts) => {
    const parts = Object.keys(queryParts).map(key => [key, queryParts[key]]);
    const str = new URLSearchParams(parts).toString();
    if (str.length > 0) {
        return `?${str}`;
    }
    return str;
};

const RoutedTab = ({ passInQuery = [], url, children }) => {
    const urlParams = useParams();
    const queryParams = useQuery();

    const queryString = {};

    passInQuery.forEach(param => {
       if (queryParams.has(param)) queryString[param] = queryParams.get(param);
    });


    const address = `${generatePath(url, urlParams)}${qs(queryString)}`;
    const active = useRouteMatch(url);

    return <Nav.Link as={Link} to={address} active={active}>{children}</Nav.Link>
};

const RepositoryTabs = () => {
    return (
        <Nav variant="tabs" defaultActiveKey="/home">
            <Nav.Item>
                <RoutedTab url="/repositories/:repoId/tree" passInQuery={['branch']}><Octicon icon={Database}/>  Objects</RoutedTab>
            </Nav.Item>
            <Nav.Item>
                <RoutedTab url="/repositories/:repoId/commits" passInQuery={['branch']}><Octicon icon={GitCommit}/>  Commits</RoutedTab>
            </Nav.Item>
            <Nav.Item>
                <RoutedTab url="/repositories/:repoId/branches"><Octicon icon={GitBranch}/> Branches</RoutedTab>
            </Nav.Item>
            <Nav.Item>
                <RoutedTab url="/repositories/:repoId/settings"><Octicon icon={Gear}/> Settings</RoutedTab>
            </Nav.Item>
        </Nav>
    );
};

const RepositoryExplorerPage = ({ repo, getRepository }) => {

    const { repoId } = useParams();
    const query = useQuery();

    useEffect(() => {
        getRepository(repoId);
    }, [getRepository, repoId]);


    if (repo.loading) {
        return (
            <div className="mt-5">
                <p>Loading...</p>
            </div>
        );
    }

    if (!!repo.error) {
        return (
            <div className="mt-5">
                <Alert variant="danger">{repo.error}</Alert>
            </div>
        );
    }

    // we have a repo object
    const branchId = query.get('branch');
    const branch = (!!branchId) ? branchId : ((!!repo.payload) ? repo.payload.default_branch : null);

    return (
        <div className="mt-5">
            <Breadcrumb>
                <Breadcrumb.Item href={`/`}>Repositories</Breadcrumb.Item>
                <Breadcrumb.Item active href={`/repositories/${repoId}`}>{repoId}</Breadcrumb.Item>
            </Breadcrumb>

            <RepositoryTabs/>

            <Switch>
                <Redirect exact from="/repositories/:repoId" to="/repositories/:repoId/tree"/>
                <Route path="/repositories/:repoId/tree">
                    <TreePage repoId={repoId} branchId={branch} path={query.get('path') || ""}/>
                </Route>
                <Route exact path="/repositories/:repoId/commits">
                    <CommitsPage repoId={repoId} branchId={branch}/>
                </Route>
            </Switch>
        </div>
    );
};

export default connect(
    ({ repositories }) => ({ repo: repositories.repo }),
    ({  getRepository })
)(RepositoryExplorerPage);