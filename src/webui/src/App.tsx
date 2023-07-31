import { useMemo } from "react";

import { Box, Toolbar } from "@mui/material";
import CssBaseline from "@mui/material/CssBaseline";
import { ThemeProvider, createTheme } from "@mui/material/styles";

import { QueryClientProvider } from "@tanstack/react-query";
import { queryClient } from "queryClient";

import { Route, Routes } from "react-router-dom";
import { PrivateRoute } from "layout/PrivateRoute";
import { privateRoutes, publicRoutes } from "layout/Routes";

import { AlertComponent } from "layout/common/AlertComponent";
import { AppBar } from "./layout/AppBar";

const mdTheme = createTheme();

export default function App() {

  const publicRoutesItems = useMemo(
    () =>
      publicRoutes.map((route) => (
        <Route key={route.link} path={route.link} element={<route.element />} />
      )),
    []
  );

  const privateRoutesItems = useMemo(
    () =>
      privateRoutes.map((route) => (
        <Route key={route.link} path={route.link} element={<PrivateRoute><route.element /></PrivateRoute>} />
      )),
    []
  );

  return (
    <QueryClientProvider client={queryClient}>
      <ThemeProvider theme={mdTheme}>
        <Box sx={{ display: "flex" }}>
          <CssBaseline />
          <AppBar />
          <Box
            component="main"
            sx={{
              flexGrow: 1,
              height: "100vh",
              overflow: "auto",
              p: 2,
              flexFlow: "column",
              display: "flex"
            }}
          >
            <Toolbar />
            <Routes>
              {/*routesItems*/}
              {publicRoutesItems}
              {privateRoutesItems}
            </Routes>
            <AlertComponent />
          </Box>
        </Box>
      </ThemeProvider>
    </QueryClientProvider>
  );
}
