// Perform sequence alignment
function performAlignment() {
    // Get the input values
    const query = document.getElementById('querySequence').value;
    const reference = document.getElementById('referenceSequence').value;
    const useParallel = document.getElementById('parallelSwitch').checked;
    const workers = parseInt(document.getElementById('workerCount').value);
    const useBatch = document.getElementById('batchSwitch').checked;
    const batchSize = parseInt(document.getElementById('batchSize').value);

    // Check if sequences are provided
    if (query.trim() === '' || reference.trim() === '') {
        alert('Please enter both query and reference sequences.');
        return;
    }

    // Create the request data
    const requestData = {
        query: query,
        reference: reference,
        useParallel: useParallel,
        workers: workers,
        useBatch: useBatch,
        batchSize: batchSize,
        generateRandom: false,
        randomLength: 0
    };

    // Show loading indicator
    document.getElementById('loadingIndicator').style.display = 'block';
    document.getElementById('resultsContainer').style.display = 'none';
    document.getElementById('batchResultsCard').style.display = 'none';

    // Send the request to the server
    fetch('/align', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(requestData)
    })
        .then(response => {
            if (!response.ok) {
                throw new Error('Server returned an error: ' + response.statusText);
            }
            return response.json();
        })
        .then(data => {
            // Hide loading indicator
            document.getElementById('loadingIndicator').style.display = 'none';
            document.getElementById('resultsContainer').style.display = 'block';

            // Display the results
            displayResults(data);

            // Update history and charts
            updateResultsHistory(data);
            updatePerformanceChart();

            // Display batch results if applicable
            if (data.batchResults && data.batchResults.length > 0) {
                displayBatchResults(data.batchResults);
            }
        })
        .catch(error => {
            // Hide loading indicator and show error
            document.getElementById('loadingIndicator').style.display = 'none';
            alert('Error performing alignment: ' + error.message);
            console.error(error);
        });
}

// Display alignment results
function displayResults(data) {
    // Update alignment score
    document.getElementById('alignmentScore').textContent = data.score;

    // Update execution time
    document.getElementById('executionTime').textContent = data.executionTimeMs.toFixed(2) + ' ms';
    document.getElementById('executionTimeDetail').textContent = data.executionTime;

    // Update execution mode
    document.getElementById('executionMode').textContent = data.isParallel ?
        `Parallel (${data.workers} workers)` :
        'Sequential';

    // Update memory usage
    document.getElementById('memoryUsage').textContent = data.memoryUsageMB + ' MB';

    // Update memory efficiency
    const bytesPerBase = data.performanceData.bytesPerBase;
    document.getElementById('memoryEfficiency').textContent =
        bytesPerBase.toFixed(2) + ' bytes/base';

    // Display aligned sequences
    const alignedQuery = data.alignedQuery;
    const alignedRef = data.alignedRef;

    document.getElementById('alignedQuery').textContent = alignedQuery;
    document.getElementById('alignedRef').textContent = alignedRef;

    // Generate and display the match line
    const matchLine = generateMatchLine(alignedQuery, alignedRef);
    document.getElementById('alignmentMatch').textContent = matchLine;
}

// Generate the match line between two aligned sequences
function generateMatchLine(seq1, seq2) {
    let matchLine = '';
    const len = Math.min(seq1.length, seq2.length);

    for (let i = 0; i < len; i++) {
        if (seq1[i] === '-' || seq2[i] === '-') {
            matchLine += ' '; // Gap
        } else if (seq1[i] === seq2[i]) {
            matchLine += '|'; // Match
        } else {
            matchLine += '.'; // Mismatch
        }
    }

    return matchLine;
}

// Display batch alignment results
function displayBatchResults(batchResults) {
    // Show the batch results card
    document.getElementById('batchResultsCard').style.display = 'block';

    // Get the table body
    const tableBody = document.getElementById('batchResultsTable').querySelector('tbody');
    tableBody.innerHTML = '';

    // Create rows for each batch result
    batchResults.forEach((result, index) => {
        const row = document.createElement('tr');

        // Add index column
        const indexCell = document.createElement('td');
        indexCell.textContent = index + 1;
        row.appendChild(indexCell);

        // Add score column
        const scoreCell = document.createElement('td');
        scoreCell.textContent = result.score;
        row.appendChild(scoreCell);

        // Add view button column
        const buttonCell = document.createElement('td');
        const viewButton = document.createElement('button');
        viewButton.className = 'btn btn-sm btn-outline-primary';
        viewButton.textContent = 'View';
        viewButton.onclick = function() {
            showBatchResultDetail(result);
        };
        buttonCell.appendChild(viewButton);
        row.appendChild(buttonCell);

        // Add the row to the table
        tableBody.appendChild(row);
    });
}

// Show detailed view of a batch result
function showBatchResultDetail(result) {
    // Set the modal content
    document.getElementById('modalAlignmentScore').textContent = result.score;
    document.getElementById('modalAlignedQuery').textContent = result.alignedQuery;
    document.getElementById('modalAlignedRef').textContent = result.alignedRef;

    // Generate and set the match line
    const matchLine = generateMatchLine(result.alignedQuery, result.alignedRef);
    document.getElementById('modalAlignmentMatch').textContent = matchLine;

    // Show the modal
    const modal = new bootstrap.Modal(document.getElementById('batchDetailModal'));
    modal.show();
}

// Update the results history for performance tracking
function updateResultsHistory(data) {
    // Create a new history entry
    const historyEntry = {
        timestamp: new Date(),
        executionTimeMs: data.executionTimeMs,
        memoryUsageMB: data.memoryUsageMB,
        sequenceLength: data.querySequence.length,
        isParallel: data.isParallel,
        workers: data.workers,
        useBatch: data.batchResults && data.batchResults.length > 0
    };

    // Add to history and limit size
    resultsHistory.push(historyEntry);
    if (resultsHistory.length > MAX_HISTORY) {
        resultsHistory.shift(); // Remove oldest entry
    }
}

// Initialize the performance chart
function initializePerformanceChart() {
    const ctx = document.getElementById('performanceChart').getContext('2d');

    performanceChart = new Chart(ctx, {
        type: 'line',
        data: {
            datasets: [
                {
                    label: 'Execution Time (ms)',
                    yAxisID: 'time',
                    borderColor: 'rgba(75, 192, 192, 1)',
                    backgroundColor: 'rgba(75, 192, 192, 0.2)',
                    data: []
                },
                {
                    label: 'Memory Usage (MB)',
                    yAxisID: 'memory',
                    borderColor: 'rgba(153, 102, 255, 1)',
                    backgroundColor: 'rgba(153, 102, 255, 0.2)',
                    data: []
                }
            ]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            scales: {
                x: {
                    type: 'linear',
                    position: 'bottom',
                    title: {
                        display: true,
                        text: 'Run #'
                    }
                },
                time: {
                    type: 'linear',
                    position: 'left',
                    title: {
                        display: true,
                        text: 'Time (ms)'
                    },
                    min: 0
                },
                memory: {
                    type: 'linear',
                    position: 'right',
                    title: {
                        display: true,
                        text: 'Memory (MB)'
                    },
                    min: 0,
                    grid: {
                        drawOnChartArea: false
                    }
                }
            },
            plugins: {
                legend: {
                    position: 'top',
                },
                tooltip: {
                    callbacks: {
                        label: function(context) {
                            const dataIndex = context.dataIndex;
                            const datasetIndex = context.datasetIndex;

                            // Get the value for the current dataset
                            const value = context.raw.y;

                            // Base label
                            let label = context.dataset.label + ': ' + value;

                            // Add information about the run
                            if (dataIndex < resultsHistory.length) {
                                const historyEntry = resultsHistory[dataIndex];
                                label += '\nLength: ' + historyEntry.sequenceLength + ' bp';
                                label += '\nMode: ' + (historyEntry.isParallel ? 'Parallel' : 'Sequential');
                                if (historyEntry.isParallel) {
                                    label += ' (' + historyEntry.workers + ' workers)';
                                }
                                label += '\nBatch: ' + (historyEntry.useBatch ? 'Yes' : 'No');
                            }

                            return label;
                        }
                    }
                }
            }
        }
    });
}

// Update the performance chart with latest results
function updatePerformanceChart() {
    // Clear existing data
    performanceChart.data.datasets[0].data = [];
    performanceChart.data.datasets[1].data = [];

    // Add data points from history
    resultsHistory.forEach((entry, index) => {
        // Add execution time data point
        performanceChart.data.datasets[0].data.push({
            x: index + 1,
            y: entry.executionTimeMs
        });

        // Add memory usage data point
        performanceChart.data.datasets[1].data.push({
            x: index + 1,
            y: entry.memoryUsageMB
        });
    });

    // Update the chart
    performanceChart.update();
}// Global variable to store the results history for performance tracking
let resultsHistory = [];
let performanceChart = null;

// Maximum number of results to store in history
const MAX_HISTORY = 10;

// Initialize the application when the DOM is loaded
document.addEventListener('DOMContentLoaded', function() {
    // Set up event listeners
    document.getElementById('generateRandomSwitch').addEventListener('change', toggleRandomControls);
    document.getElementById('parallelSwitch').addEventListener('change', toggleParallelControls);
    document.getElementById('batchSwitch').addEventListener('change', toggleBatchControls);
    document.getElementById('generateBtn').addEventListener('click', generateRandomSequences);
    document.getElementById('alignBtn').addEventListener('click', performAlignment);

    // Initialize controls
    toggleRandomControls();
    toggleParallelControls();
    toggleBatchControls();

    // Initialize performance chart
    initializePerformanceChart();

    // Add example sequences
    document.getElementById('querySequence').value = 'GATTACACGGTAGATCAGATAGATACACGTTCGATCGACTAGCTAGATA';
    document.getElementById('referenceSequence').value = 'GATACACTGTAGATCTGATAGATACACTTTCGATCGACTAGCCAGATA';
});

// Toggle controls for random sequence generation
function toggleRandomControls() {
    const isChecked = document.getElementById('generateRandomSwitch').checked;
    document.getElementById('randomControls').style.display = isChecked ? 'block' : 'none';
    document.getElementById('sequenceInputs').style.display = isChecked ? 'none' : 'block';
}

// Toggle controls for parallel execution
function toggleParallelControls() {
    const isChecked = document.getElementById('parallelSwitch').checked;
    document.getElementById('parallelControls').style.display = isChecked ? 'block' : 'none';
}

// Toggle controls for batch processing
function toggleBatchControls() {
    const isChecked = document.getElementById('batchSwitch').checked;
    document.getElementById('batchControls').style.display = isChecked ? 'block' : 'none';
    document.getElementById('batchResultsCard').style.display = 'none';
}

// Generate random DNA sequences
function generateRandomSequences() {
    const length = parseInt(document.getElementById('randomLength').value);
    if (isNaN(length) || length < 10 || length > 10000) {
        alert('Please enter a valid sequence length between 10 and 10000.');
        return;
    }

    // Generate random sequences
    document.getElementById('querySequence').value = generateRandomDNA(length);
    document.getElementById('referenceSequence').value = generateRandomDNA(length);

    // Toggle the display to show the sequences
    document.getElementById('generateRandomSwitch').checked = false;
    toggleRandomControls();
}

// Generate a random DNA sequence
function generateRandomDNA(length) {
    const bases = ['A', 'C', 'G', 'T'];
    let sequence = '';
    for (let i = 0; i < length; i++) {
        sequence += bases[Math.floor(Math.random() * bases.length)];
    }
    return sequence;
}