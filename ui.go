package main

import (
	"fmt"
	"net/http"
)

// ServeDashboardHandler outputs a native dashboard interface directly to browser viewports
func ServeDashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	
	fmt.Fprint(w, `
	<!DOCTYPE html>
	<html lang="en">
	<head>
		<meta charset="UTF-8">
		<title>Failed Transaction & Dispute Tracker</title>
		<style>
			body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; background: #7e6baa; margin: 0; padding: 40px; color: #333; }
			.container { max-width: 900px; margin: 0 auto; }
			header { background: #fff; padding: 24px; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); margin-bottom: 30px; }
			h1 { margin: 0; color: #0d6efd; font-size: 24px; }
			.card { background: #fff; padding: 24px; border-radius: 8px; box-shadow: 0 4px 12px rgba(0,0,0,0.05); margin-bottom: 30px; }
			.form-group { margin-bottom: 15px; }
			label { display: block; margin-bottom: 6px; font-weight: 600; font-size: 14px; }
			input, select { width: 100%; padding: 10px; border: 1px solid #ced4da; border-radius: 4px; box-sizing: border-box; }
			button { background: #0d6efd; color: #fff; border: none; padding: 12px 20px; font-weight: 600; border-radius: 4px; cursor: pointer; width: 100%; transition: 0.2s; }
			button:hover { background: #0b5ed7; }
			table { width: 100%; border-collapse: collapse; margin-top: 15px; }
			th, td { padding: 12px; text-align: left; border-bottom: 1px solid #dee2e6; font-size: 14px; }
			th { background: #f8f9fa; font-weight: 600; }
			.badge { padding: 4px 8px; border-radius: 4px; font-size: 12px; font-weight: 600; }
			.badge-danger { background: #f8d7da; color: #842029; }
			.badge-success { background: #d1e7dd; color: #0f5132; }
			.esc-btn { background: #dc3545; color: white; padding: 6px 12px; border-radius: 4px; font-size: 12px; text-decoration: none; display: inline-block; text-align: center; width: auto;}
			.esc-btn:hover { background: #bb2d3b; }
		</style>
	</head>
	<body>
		<div class="container">
			<header>
				<h1>🇳🇬 NGR Bank Reversal Tracker</h1>
				<p style="margin: 4px 0 0 0; color: #666; font-size: 14px;">Log transfers and generate legal templates for delayed funds.</p>
			</header>

			<div class="card">
				<h3 style="margin-top:0;">Log Failed Transaction</h3>
				<form id="txForm">
					<div class="form-group">
						<label>Sender Bank (Where money left - Includes Digital Banks & Fintechs)</label>
						<select id="sender_bank" required>
							<option value="">-- Select Source Bank --</option>
						</select>
					</div>
					<div class="form-group">
						<label>Receiver Bank Name (Destination - Includes Digital Banks & Fintechs)</label>
						<select id="receiver_bank" required>
							<option value="">-- Select Destination Bank --</option>
						</select>
					</div>
					<div class="form-group">
						<label>Transaction Value Amount (₦)</label>
						<input type="number" step="0.01" id="amount" placeholder="e.g. 5000" required>
					</div>
					<div class="form-group">
						<label>NIBSS 30-Digit Session ID</label>
						<input type="text" id="session_id" placeholder="Enter the 30-digit reference code" maxlength="30" required>
					</div>
					<div class="form-group">
						<label>Transaction Timestamp</label>
						<input type="datetime-local" id="transfer_date" required>
					</div>
					<button type="submit">Securely Log Dispute Record</button>
				</form>
			</div>

			<div class="card">
				<h3 style="margin-top:0;">Active Logs Database</h3>
				<div style="overflow-x:auto;">
					<table>
						<thead>
							<tr>
								<th>ID</th>
								<th>Route</th>
								<th>Amount</th>
								<th>Session ID</th>
								<th>Timeline Status</th>
								<th>Action</th>
							</tr>
						</thead>
						<tbody id="logsTable"></tbody>
					</table>
				</div>
			</div>
		</div>

		<script>
			// Master collection of ALL traditional commercial banks, viral digital neobanks, and telecom PSBs in Nigeria
			const nigeriaBanks = [
				"9PSB", "Access Bank", "Brass", "Carbon", "FairMoney", "FCMB", 
				"Fidelity Bank", "First Bank", "GoMoney", "GTBank", "Heritage Bank", 
				"Keystone Bank", "Kuda Bank", "MoMo PSB", "Moniepoint", "OPay", 
				"Optimus Bank", "PalmPay", "PiggyVest", "PocketApp", "Polaris Bank", 
				"PremiumTrust Bank", "Providus Bank", "Rubies Bank", "SmartCash PSB", 
				"Sparkle", "Stanbic IBTC", "Standard Chartered", "Sterling Bank", 
				"SunTrust Bank", "UBA", "Union Bank", "Unity Bank", "VFD Microfinance", 
				"Wema Bank", "Zenith Bank"
			];

			// Programmatically generate option items for BOTH lists simultaneously 
			const populateDropdowns = () => {
				const senderSelect = document.getElementById('sender_bank');
				const receiverSelect = document.getElementById('receiver_bank');
				
				nigeriaBanks.forEach(bank => {
					// Add to Sender Select menu
					const opt1 = document.createElement('option');
					opt1.value = bank;
					opt1.textContent = bank;
					senderSelect.appendChild(opt1);

					// Add to Receiver Select menu
					const opt2 = document.createElement('option');
					opt2.value = bank;
					opt2.textContent = bank;
					receiverSelect.appendChild(opt2);
				});
			};

			const loadLogs = async () => {
				const res = await fetch('/api/transfers');
				const data = await res.json();
				const tbody = document.getElementById('logsTable');
				tbody.innerHTML = '';
				data.forEach(log => {
					const row = document.createElement('tr');
					row.innerHTML = '<td>' + log.id + '</td>' +
						'<td><b>' + log.sender_bank + '</b> ➔ ' + log.receiver_bank + '</td>' +
						'<td>₦' + log.amount.toLocaleString() + '</td>' +
						'<td style="font-family:monospace; font-size:12px;">' + log.session_id + '</td>' +
						'<td>' + (log.can_escalate ? '<span class="badge badge-danger">Over 48h - Unresolved</span>' : '<span class="badge badge-success">Processing Window</span>') + '</td>' +
						'<td>' + (log.can_escalate ? '<button class="esc-btn" onclick="escalate(' + log.id + ')">Generate Email</button>' : '<span style="color:#aaa; font-size:12px;">Locked</span>') + '</td>';
					tbody.appendChild(row);
				});
			};

			document.getElementById('txForm').addEventListener('submit', async (e) => {
				e.preventDefault();
				const payload = {
					sender_bank: document.getElementById('sender_bank').value,
					receiver_bank: document.getElementById('receiver_bank').value,
					amount: parseFloat(document.getElementById('amount').value),
					session_id: document.getElementById('session_id').value,
					transfer_date: new Date(document.getElementById('transfer_date').value).toISOString()
				};
				
				const res = await fetch('/api/transfers', { 
					method: 'POST', 
					headers: {'Content-Type': 'application/json'}, 
					body: JSON.stringify(payload) 
				});
				
				if (!res.ok) {
					const errMsg = await res.text();
					alert(errMsg);
					return;
				}
				
				document.getElementById('txForm').reset();
				loadLogs();
			});

			const escalate = async (id) => {
				const res = await fetch('/api/escalate?id=' + id, { method: 'POST' });
				const data = await res.json();
				alert("E-mail Compiled! Check your Codespaces Terminal output stream console to read the full letter template formatted for: " + data.dispatched_to);
			};

			// Initialize layouts on system startup
			populateDropdowns();
			loadLogs();
		</script>
	</body>
	</html>
	`)
}
