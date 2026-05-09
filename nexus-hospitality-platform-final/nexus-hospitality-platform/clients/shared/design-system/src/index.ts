import React from 'react';

export const DesignSystemProvider: React.FC<{
  theme?: string;
  children?: React.ReactNode;
}> = ({ theme = 'enterprise', children }) => {
  return React.createElement('div', { 'data-theme': theme }, children);
};

export const LoadingScreen: React.FC<{ fullScreen?: boolean }> = ({ fullScreen }) => (
  React.createElement('div', {
    style: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      height: fullScreen ? '100vh' : '100%',
    }
  }, 'Loading...')
);
