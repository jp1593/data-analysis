import numpy as np
import matplotlib.pyplot as plt
from scipy.io import loadmat
from scipy.sparse.csgraph import dijkstra
from sklearn.neighbors import NearestNeighbors
from sklearn.decomposition import PCA
from scipy.sparse import lil_matrix

# Settings
K_NEIGHBORS = 7
OUTPUT_DIMS = 2
SUBSET_SIZE = 1000

mat_data = loadmat("isomap.mat")
X = mat_data['images']

if X.shape[0] > X.shape[1]:
    X = X.T

# Subset take
np.random.seed(42)
if SUBSET_SIZE < X.shape[0]:
    idx = np.random.choice(X.shape[0], SUBSET_SIZE, replace=False)
    X = X[idx]

# Nearest Neighbors Graph
nbrs = NearestNeighbors(n_neighbors=K_NEIGHBORS).fit(X)
distances, indices = nbrs.kneighbors(X)

N = X.shape[0]
graph = lil_matrix((N, N))
for i in range(N):
    for j, dist in zip(indices[i], distances[i]):
        graph[i, j] = dist
        graph[j, i] = dist

# Geodetic distances
geo_distances = dijkstra(csgraph=graph, directed=False)

# MDS (ISOMAP)
n = geo_distances.shape[0]
H = np.eye(n) - np.ones((n, n)) / n
K_matrix = -0.5 * H @ (geo_distances ** 2) @ H

eigvals, eigvecs = np.linalg.eigh(K_matrix)
idx = np.argsort(eigvals)[::-1]
eigvals = np.abs(eigvals)
eigvals_pos = eigvals[:OUTPUT_DIMS]
eigvecs_pos = eigvecs[:, :OUTPUT_DIMS]
Y_isomap = eigvecs_pos * np.sqrt(eigvals_pos)

# PCA
pca = PCA(n_components=OUTPUT_DIMS)
Y_pca = pca.fit_transform(X)

# Plot of ISOMAP vs PCA
plt.figure(figsize=(16,7))

# ISOMAP
plt.subplot(1, 2, 1)
scatter1 = plt.scatter(Y_isomap[:,0], Y_isomap[:,1], c=np.arange(len(Y_isomap)),
                       cmap='viridis', s=50, alpha=0.8, edgecolors='k')
plt.title("ISOMAP – Reducción no lineal a 2D", fontsize=16, fontweight='bold')
plt.xlabel("Coordenada 1 (distancia geodésica preservada)", fontsize=12)
plt.ylabel("Coordenada 2 (distancia geodésica preservada)", fontsize=12)
plt.grid(True)
plt.axis('equal')

# PCA
plt.subplot(1, 2, 2)
scatter2 = plt.scatter(Y_pca[:,0], Y_pca[:,1], c=np.arange(len(Y_pca)),
                       cmap='plasma', s=50, alpha=0.8, edgecolors='k')
plt.title("PCA – Reducción lineal a 2D", fontsize=16, fontweight='bold')
plt.xlabel("Componente 1 (mayor varianza)", fontsize=12)
plt.ylabel("Componente 2 (segunda mayor varianza)", fontsize=12)
plt.grid(True)
plt.axis('equal')

plt.tight_layout()
plt.show()
