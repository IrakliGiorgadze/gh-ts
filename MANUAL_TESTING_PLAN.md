# Manual Testing Plan - IT Ticketing System

## Test Environment Setup
- **Base URL**: http://localhost:3000
- **API URL**: http://localhost:8080
- **Test Users**: 
  - Admin: ika.giorgadze@gmail.com / @dm!n123
  - End User: (create via registration)
  - Agent: (create via admin panel)

---

## 1. AUTHENTICATION MODULE

### 1.1 Login Page (`/pages/login.html`)

#### Test Case 1.1.1: Valid Login
**Steps:**
1. Navigate to http://localhost:3000/pages/login.html
2. Enter valid email: `ika.giorgadze@gmail.com`
3. Enter valid password: `@dm!n123`
4. Click "Login" button

**Expected Results:**
- Login succeeds
- Redirects to `/index.html`
- Session cookie is set
- User remains authenticated after page refresh

#### Test Case 1.1.2: Invalid Credentials
**Steps:**
1. Navigate to login page
2. Enter invalid email: `wrong@example.com`
3. Enter invalid password: `wrongpass`
4. Click "Login"

**Expected Results:**
- Error message displayed: "Invalid credentials"
- User stays on login page
- No session cookie set

#### Test Case 1.1.3: Empty Fields
**Steps:**
1. Navigate to login page
2. Leave email empty
3. Leave password empty
4. Click "Login"

**Expected Results:**
- Error message: "Email and password are required"
- Form does not submit

#### Test Case 1.1.4: Login with ?next= Parameter
**Steps:**
1. Navigate to http://localhost:3000/pages/login.html?next=%2Fpages%2Ftickets.html
2. Enter valid credentials
3. Click "Login"

**Expected Results:**
- Login succeeds
- Redirects to `/pages/tickets.html` (the next parameter)

#### Test Case 1.1.5: Already Logged In User
**Steps:**
1. Log in successfully
2. Navigate to `/pages/login.html` while still logged in

**Expected Results:**
- Should redirect to home page (or handle gracefully)
- Should not show login form

---

### 1.2 Registration Page (`/pages/register.html`)

#### Test Case 1.2.1: Valid Registration
**Steps:**
1. Navigate to http://localhost:3000/pages/register.html
2. Enter email: `newuser@example.com`
3. Enter name: `New User`
4. Enter password: `Password123` (meets requirements)
5. Click "Register"

**Expected Results:**
- Registration succeeds
- Success message displayed
- Redirects to login page after 800ms
- User role is set to "end_user"

#### Test Case 1.2.2: Invalid Email Format
**Steps:**
1. Navigate to register page
2. Enter invalid email: `notanemail`
3. Enter name and valid password
4. Click "Register"

**Expected Results:**
- Error message: "invalid email format"
- Registration fails

#### Test Case 1.2.3: Weak Password
**Steps:**
1. Navigate to register page
2. Enter valid email and name
3. Enter weak password: `weak` (less than 8 chars)
4. Click "Register"

**Expected Results:**
- Error message about password requirements
- Registration fails

#### Test Case 1.2.4: Password Requirements
**Test each requirement:**
- Password without uppercase: `password123` → Should fail
- Password without lowercase: `PASSWORD123` → Should fail
- Password without number: `Password` → Should fail
- Valid password: `Password123` → Should succeed

#### Test Case 1.2.5: Duplicate Email
**Steps:**
1. Register a user with email `test@example.com`
2. Try to register again with same email

**Expected Results:**
- Error message: "resource already exists" (prod) or detailed error (dev)
- Registration fails

#### Test Case 1.2.6: Empty Fields
**Steps:**
1. Navigate to register page
2. Leave fields empty
3. Click "Register"

**Expected Results:**
- Validation errors displayed
- Registration fails

---

### 1.3 Logout

#### Test Case 1.3.1: Logout from Navbar
**Steps:**
1. Log in successfully
2. Click "Logout" button in navbar
3. Verify logout

**Expected Results:**
- Session cookie cleared
- Redirects to login page
- Cannot access protected pages

#### Test Case 1.3.2: Logout and Access Protected Page
**Steps:**
1. Log in
2. Log out
3. Try to access `/index.html`

**Expected Results:**
- Redirects to login with `?next=` parameter
- Cannot see dashboard content

---

## 2. DASHBOARD / HOME PAGE (`/index.html`)

### 2.1 Unauthenticated Access

#### Test Case 2.1.1: Access Without Login
**Steps:**
1. Clear cookies/session
2. Navigate to http://localhost:3000/index.html

**Expected Results:**
- Page content is hidden (display: none)
- Immediately redirects to `/pages/login.html?next=%2Findex.html`
- No dashboard content visible (KPIs, tickets, statistics)

#### Test Case 2.1.2: Direct URL Access
**Steps:**
1. While logged out, directly type `/index.html` in browser
2. Observe page behavior

**Expected Results:**
- Redirects to login before any content loads
- No flash of dashboard content

---

### 2.2 Authenticated Access

#### Test Case 2.2.1: Dashboard Load
**Steps:**
1. Log in as admin
2. Navigate to `/index.html`

**Expected Results:**
- Dashboard loads successfully
- Shows three KPI cards:
  - Open Tickets (count)
  - SLA at Risk (count)
  - Resolved (7d) (count)
- Shows tickets table with columns: ID, Title, Priority, Status, Assignee, Updated
- Shows search input and status filter
- Shows "New Ticket" button

#### Test Case 2.2.2: KPI Statistics
**Steps:**
1. Log in
2. View dashboard
3. Check KPI values

**Expected Results:**
- KPIs display actual counts from API
- Values update correctly
- If API fails, shows 0 (graceful degradation)

#### Test Case 2.2.3: Tickets Table
**Steps:**
1. Log in
2. View dashboard
3. Check tickets table

**Expected Results:**
- Table displays tickets from API
- Each ticket shows: ID (or alias), Title (clickable link), Priority, Status (with badge), Assignee, Updated date
- If no tickets, shows "No tickets found"

#### Test Case 2.2.4: Search Functionality
**Steps:**
1. Log in
2. Type search query in search box
3. Observe results

**Expected Results:**
- Filters tickets by title or description
- Updates table in real-time
- Handles empty search (shows all tickets)

#### Test Case 2.2.5: Status Filter
**Steps:**
1. Log in
2. Select status from dropdown (e.g., "Open")
3. Observe filtered results

**Expected Results:**
- Filters tickets by selected status
- Updates table immediately
- "All Statuses" shows all tickets

#### Test Case 2.2.6: Admin Quick Access (Admin Only)
**Steps:**
1. Log in as admin
2. View dashboard

**Expected Results:**
- Shows "Admin Panel" card with "Manage Users" button
- Card visible only to admin users

#### Test Case 2.2.7: Admin Quick Access (Non-Admin)
**Steps:**
1. Log in as end_user or agent
2. View dashboard

**Expected Results:**
- Admin Panel card is hidden
- Only regular dashboard content visible

#### Test Case 2.2.8: Page Refresh
**Steps:**
1. Log in
2. View dashboard
3. Refresh page (F5)

**Expected Results:**
- Page reloads
- User remains authenticated
- Dashboard content visible
- No redirect to login

---

## 3. TICKET MANAGEMENT MODULE

### 3.1 Create Ticket (`/pages/create-ticket.html`)

#### Test Case 3.1.1: Create Ticket - Valid Input
**Steps:**
1. Log in
2. Navigate to `/pages/create-ticket.html` or click "New Ticket" button
3. Fill in:
   - Title: "Test Ticket"
   - Description: "This is a test ticket description"
   - Category: Select from dropdown
   - Priority: Select from dropdown
   - Department: "IT Support"
4. Click "Create Ticket"

**Expected Results:**
- Ticket created successfully
- Success message displayed
- Redirects to ticket detail page
- Ticket appears in tickets list

#### Test Case 3.1.2: Create Ticket - Empty Title
**Steps:**
1. Log in
2. Navigate to create ticket page
3. Leave title empty
4. Fill other fields
5. Click "Create Ticket"

**Expected Results:**
- Error message: "title is required"
- Ticket not created

#### Test Case 3.1.3: Create Ticket - Title Too Long
**Steps:**
1. Log in
2. Navigate to create ticket page
3. Enter title longer than 200 characters
4. Click "Create Ticket"

**Expected Results:**
- Error message: "title too long (max 200 characters)"
- Ticket not created

#### Test Case 3.1.4: Create Ticket - Description Too Long
**Steps:**
1. Log in
2. Navigate to create ticket page
3. Enter description longer than 5000 characters
4. Click "Create Ticket"

**Expected Results:**
- Error message: "description too long (max 5000 characters)"
- Ticket not created

#### Test Case 3.1.5: Create Ticket - Department Too Long
**Steps:**
1. Log in
2. Navigate to create ticket page
3. Enter department longer than 100 characters
4. Click "Create Ticket"

**Expected Results:**
- Error message: "department too long (max 100 characters)"
- Ticket not created

#### Test Case 3.1.6: Create Ticket - Unauthenticated
**Steps:**
1. Log out
2. Try to access `/pages/create-ticket.html`

**Expected Results:**
- Redirects to login page
- Cannot create ticket

---

### 3.2 Ticket List (`/pages/tickets.html`)

#### Test Case 3.2.1: View Tickets List
**Steps:**
1. Log in
2. Navigate to `/pages/tickets.html`

**Expected Results:**
- Displays list of all tickets
- Shows ticket details: ID, Title, Priority, Status, Assignee, Updated
- Tickets are clickable (link to detail page)

#### Test Case 3.2.2: Search Tickets
**Steps:**
1. Log in
2. Navigate to tickets page
3. Enter search term in search box

**Expected Results:**
- Filters tickets by search term
- Updates list in real-time
- Search works on title and description

#### Test Case 3.2.3: Filter by Status
**Steps:**
1. Log in
2. Navigate to tickets page
3. Select status from filter dropdown

**Expected Results:**
- Filters tickets by selected status
- Updates list immediately

#### Test Case 3.2.4: Pagination (if implemented)
**Steps:**
1. Log in
2. Navigate to tickets page
3. Check if pagination controls exist
4. Navigate through pages

**Expected Results:**
- Pagination works correctly
- Page numbers update
- Tickets load for each page

---

### 3.3 Ticket Detail (`/pages/ticket-detail.html`)

#### Test Case 3.3.1: View Ticket Details
**Steps:**
1. Log in
2. Navigate to a ticket detail page (e.g., `/pages/ticket-detail.html?id=<ticket-id>`)

**Expected Results:**
- Displays full ticket information:
  - Title
  - Description
  - Status
  - Priority
  - Category
  - Assignee
  - Created by
  - Created date
  - Updated date
- Shows comments section
- Shows update form (if user has permission)

#### Test Case 3.3.2: Update Ticket Status
**Steps:**
1. Log in as agent or admin
2. Open a ticket
3. Change status from dropdown
4. Click "Update" or "Save"

**Expected Results:**
- Status updates successfully
- Page refreshes with new status
- Status change reflected in tickets list

#### Test Case 3.3.3: Update Ticket Priority
**Steps:**
1. Log in as agent or admin
2. Open a ticket
3. Change priority
4. Save changes

**Expected Results:**
- Priority updates successfully
- Changes reflected immediately

#### Test Case 3.3.4: Assign Ticket
**Steps:**
1. Log in as agent or admin
2. Open an unassigned ticket
3. Select assignee from dropdown
4. Save changes

**Expected Results:**
- Ticket assigned successfully
- Assignee name displayed in ticket
- Assignment reflected in tickets list

#### Test Case 3.3.5: Add Comment
**Steps:**
1. Log in
2. Open a ticket
3. Enter comment in comment box
4. Click "Add Comment" or "Submit"

**Expected Results:**
- Comment added successfully
- Comment appears in comments section
- Shows comment author and timestamp
- Comments ordered chronologically

#### Test Case 3.3.6: Comment Too Long
**Steps:**
1. Log in
2. Open a ticket
3. Enter comment longer than 2000 characters
4. Submit

**Expected Results:**
- Error message: "comment too long (max 2000 characters)"
- Comment not added

#### Test Case 3.3.7: Invalid Ticket ID
**Steps:**
1. Log in
2. Navigate to `/pages/ticket-detail.html?id=invalid-id`

**Expected Results:**
- Error message displayed
- Handles gracefully (shows error or redirects)

#### Test Case 3.3.8: Non-Existent Ticket
**Steps:**
1. Log in
2. Navigate to `/pages/ticket-detail.html?id=<non-existent-uuid>`

**Expected Results:**
- Error message: "ticket not found" or similar
- Handles gracefully

#### Test Case 3.3.9: Unauthenticated Access
**Steps:**
1. Log out
2. Try to access ticket detail page

**Expected Results:**
- Redirects to login page
- Cannot view ticket

---

## 4. USER MANAGEMENT MODULE (Admin Only)

### 4.1 User List (`/pages/admin/users.html`)

#### Test Case 4.1.1: View Users List (Admin)
**Steps:**
1. Log in as admin
2. Navigate to `/pages/admin/users.html` or click "Users" in navbar

**Expected Results:**
- Displays list of all users
- Shows user details: Email, Name, Role, Status (Active/Inactive)
- Shows "Create User" form

#### Test Case 4.1.2: View Users List (Non-Admin)
**Steps:**
1. Log in as end_user or agent
2. Try to access `/pages/admin/users.html`

**Expected Results:**
- Should redirect or show 403 error
- Cannot access user management

#### Test Case 4.1.3: Search Users
**Steps:**
1. Log in as admin
2. Navigate to users page
3. Enter search term in search box

**Expected Results:**
- Filters users by email or name
- Updates list in real-time

#### Test Case 4.1.4: Filter by Role
**Steps:**
1. Log in as admin
2. Navigate to users page
3. Select role from filter dropdown

**Expected Results:**
- Filters users by selected role
- Updates list immediately

---

### 4.2 Create User (Admin Only)

#### Test Case 4.2.1: Create User - Valid Input
**Steps:**
1. Log in as admin
2. Navigate to users page
3. Fill in create user form:
   - Email: `newuser@example.com`
   - Name: `New User`
   - Password: `Password123`
   - Role: Select from dropdown (admin, agent, end_user)
4. Click "Create User"

**Expected Results:**
- User created successfully
- Success message displayed
- User appears in users list
- User can log in with created credentials

#### Test Case 4.2.2: Create User - Invalid Email
**Steps:**
1. Log in as admin
2. Navigate to users page
3. Enter invalid email format
4. Fill other fields
5. Click "Create User"

**Expected Results:**
- Error message: "invalid email format"
- User not created

#### Test Case 4.2.3: Create User - Weak Password
**Steps:**
1. Log in as admin
2. Navigate to users page
3. Enter weak password
4. Click "Create User"

**Expected Results:**
- Error message about password requirements
- User not created

#### Test Case 4.2.4: Create User - Duplicate Email
**Steps:**
1. Log in as admin
2. Create a user with email `test@example.com`
3. Try to create another user with same email

**Expected Results:**
- Error message: "resource already exists"
- User not created

#### Test Case 4.2.5: Create User - Name Too Long
**Steps:**
1. Log in as admin
2. Navigate to users page
3. Enter name longer than 100 characters
4. Click "Create User"

**Expected Results:**
- Error message: "name too long (max 100 characters)"
- User not created

#### Test Case 4.2.6: Create User - All Roles
**Test creating users with each role:**
- Admin user
- Agent user
- End user

**Expected Results:**
- Each role can be created successfully
- Created users have correct role permissions

---

### 4.3 Update User

#### Test Case 4.3.1: Update User Name
**Steps:**
1. Log in as admin
2. Navigate to users page
3. Find a user and click "Edit" or update name
4. Change name
5. Save

**Expected Results:**
- Name updates successfully
- Changes reflected in users list

#### Test Case 4.3.2: Update User Role
**Steps:**
1. Log in as admin
2. Navigate to users page
3. Change user's role
4. Save

**Expected Results:**
- Role updates successfully
- User permissions change accordingly

#### Test Case 4.3.3: Deactivate User
**Steps:**
1. Log in as admin
2. Navigate to users page
3. Deactivate a user (set active to false)

**Expected Results:**
- User deactivated
- User cannot log in
- Status shows as inactive

#### Test Case 4.3.4: Reactivate User
**Steps:**
1. Log in as admin
2. Navigate to users page
3. Reactivate a deactivated user

**Expected Results:**
- User reactivated
- User can log in again
- Status shows as active

---

## 5. REPORTS MODULE (`/pages/reports.html`)

#### Test Case 5.1: View Reports
**Steps:**
1. Log in
2. Navigate to `/pages/reports.html`

**Expected Results:**
- Displays reports dashboard
- Shows statistics by status
- Shows statistics by priority
- Shows statistics by category
- Charts/graphs render correctly (if implemented)

#### Test Case 5.2: Reports - Unauthenticated
**Steps:**
1. Log out
2. Try to access `/pages/reports.html`

**Expected Results:**
- Redirects to login page
- Cannot view reports

#### Test Case 5.3: Reports Data Accuracy
**Steps:**
1. Log in
2. View reports
3. Compare with actual ticket data

**Expected Results:**
- Report numbers match actual ticket counts
- Statistics are accurate

---

## 6. NAVIGATION & UI COMPONENTS

### 6.1 Navbar

#### Test Case 6.1.1: Navbar Display
**Steps:**
1. Log in
2. Check navbar on all pages

**Expected Results:**
- Navbar visible on all pages
- Shows user name/email
- Shows "Logout" button
- Shows "Users" link (admin only)

#### Test Case 6.1.2: Navbar - Admin User
**Steps:**
1. Log in as admin
2. Check navbar

**Expected Results:**
- Shows "Users" link
- Shows admin-specific options (if any)

#### Test Case 6.1.3: Navbar - Non-Admin User
**Steps:**
1. Log in as end_user or agent
2. Check navbar

**Expected Results:**
- "Users" link is hidden
- Only standard navigation visible

#### Test Case 6.1.4: Navbar - Logged Out
**Steps:**
1. Log out
2. Check navbar on login/register pages

**Expected Results:**
- Navbar may show "Login" link or minimal content
- Appropriate for unauthenticated state

---

## 7. SECURITY & ERROR HANDLING

### 7.1 Input Validation

#### Test Case 7.1.1: XSS Attempt
**Steps:**
1. Log in
2. Try to enter `<script>alert('xss')</script>` in any text field
3. Submit form

**Expected Results:**
- Input is sanitized
- Script tags are escaped
- No script execution

#### Test Case 7.1.2: SQL Injection Attempt
**Steps:**
1. Log in
2. Try to enter SQL injection payloads in search fields
3. Submit

**Expected Results:**
- Input is sanitized
- No SQL errors
- System handles gracefully

#### Test Case 7.1.3: Path Traversal
**Steps:**
1. Log in
2. Try to access `/pages/../../../etc/passwd` or similar

**Expected Results:**
- Request blocked or handled safely
- No file system access

---

### 7.2 Rate Limiting

#### Test Case 7.2.1: Login Rate Limit
**Steps:**
1. Try to log in 11 times rapidly with wrong credentials

**Expected Results:**
- After 10 attempts, get 429 Too Many Requests
- Rate limit message displayed
- Must wait before trying again

#### Test Case 7.2.2: API Rate Limit
**Steps:**
1. Log in
2. Make 201 API requests rapidly

**Expected Results:**
- After 200 requests, get 429 error
- Rate limit enforced

---

### 7.3 Session Management

#### Test Case 7.3.1: Session Expiry
**Steps:**
1. Log in
2. Wait for session to expire (24 hours)
3. Try to access protected page

**Expected Results:**
- Session expires
- Redirects to login
- Must log in again

#### Test Case 7.3.2: Multiple Tabs
**Steps:**
1. Log in
2. Open multiple tabs
3. Log out in one tab
4. Try to use other tabs

**Expected Results:**
- Other tabs detect logout
- Redirect to login or show appropriate message

#### Test Case 7.3.3: Cookie Security
**Steps:**
1. Log in
2. Check browser DevTools → Application → Cookies

**Expected Results:**
- Session cookie has HttpOnly flag
- Session cookie has SameSite=Lax
- Secure flag set based on environment (prod=true, dev=false)

---

### 7.4 Error Handling

#### Test Case 7.4.1: Network Error
**Steps:**
1. Log in
2. Disconnect network
3. Try to perform actions

**Expected Results:**
- Error messages displayed
- Graceful degradation
- No crashes

#### Test Case 7.4.2: Server Error (500)
**Steps:**
1. Log in
2. Trigger a server error (if possible)
3. Observe error handling

**Expected Results:**
- Error message displayed
- In production: Generic error message
- In development: Detailed error message
- Application remains stable

#### Test Case 7.4.3: Not Found (404)
**Steps:**
1. Log in
2. Navigate to non-existent page: `/pages/nonexistent.html`

**Expected Results:**
- 404 error or redirect
- Handled gracefully

---

## 8. CROSS-BROWSER TESTING

#### Test Case 8.1: Chrome
**Steps:**
1. Test all major features in Chrome

**Expected Results:**
- All features work correctly
- UI renders properly

#### Test Case 8.2: Firefox
**Steps:**
1. Test all major features in Firefox

**Expected Results:**
- All features work correctly
- UI renders properly

#### Test Case 8.3: Safari
**Steps:**
1. Test all major features in Safari

**Expected Results:**
- All features work correctly
- UI renders properly

#### Test Case 8.4: Edge
**Steps:**
1. Test all major features in Edge

**Expected Results:**
- All features work correctly
- UI renders properly

---

## 9. RESPONSIVE DESIGN

#### Test Case 9.1: Mobile View
**Steps:**
1. Open application on mobile device or resize browser to mobile size
2. Test all pages

**Expected Results:**
- Layout adapts to mobile
- All features accessible
- Text readable
- Buttons clickable

#### Test Case 9.2: Tablet View
**Steps:**
1. Resize browser to tablet size
2. Test all pages

**Expected Results:**
- Layout adapts appropriately
- Features work correctly

#### Test Case 9.3: Desktop View
**Steps:**
1. Use full desktop browser
2. Test all pages

**Expected Results:**
- Optimal layout displayed
- All features accessible

---

## 10. EDGE CASES & BOUNDARY CONDITIONS

#### Test Case 10.1: Very Long Inputs
**Steps:**
1. Test with inputs at maximum length limits
2. Test with inputs just over maximum length

**Expected Results:**
- Maximum length inputs accepted
- Over maximum length rejected with error

#### Test Case 10.2: Special Characters
**Steps:**
1. Test inputs with special characters: `!@#$%^&*()`
2. Test with unicode characters
3. Test with emojis

**Expected Results:**
- Handled correctly
- No errors or crashes

#### Test Case 10.3: Empty Database
**Steps:**
1. Test with empty database (no tickets, no users except admin)

**Expected Results:**
- Application handles gracefully
- Shows appropriate empty states
- No errors

#### Test Case 10.4: Large Dataset
**Steps:**
1. Create many tickets (100+)
2. Test pagination and performance

**Expected Results:**
- Application performs well
- Pagination works correctly
- No performance issues

---

## 11. INTEGRATION TESTING

#### Test Case 11.1: End-to-End Workflow
**Steps:**
1. Register new user
2. Log in
3. Create ticket
4. View ticket
5. Add comment
5. Update ticket status
6. View in tickets list
7. Search for ticket
8. Log out

**Expected Results:**
- Complete workflow works seamlessly
- All steps execute correctly
- Data persists correctly

#### Test Case 11.2: Admin Workflow
**Steps:**
1. Log in as admin
2. Create new user (agent)
3. Assign ticket to agent
4. View reports
5. Manage users
6. Log out

**Expected Results:**
- All admin functions work
- Permissions enforced correctly

---

## 12. PERFORMANCE TESTING

#### Test Case 12.1: Page Load Time
**Steps:**
1. Measure time to load each page
2. Check with browser DevTools

**Expected Results:**
- Pages load within acceptable time (< 2 seconds)
- No significant delays

#### Test Case 12.2: API Response Time
**Steps:**
1. Monitor API calls in Network tab
2. Check response times

**Expected Results:**
- API responses are fast (< 500ms for most requests)
- No timeouts

#### Test Case 12.3: Concurrent Users
**Steps:**
1. Test with multiple users logged in simultaneously
2. Perform actions concurrently

**Expected Results:**
- System handles concurrent requests
- No conflicts or data corruption

---

## TESTING CHECKLIST

### Pre-Testing Setup
- [ ] Docker containers running
- [ ] Database initialized
- [ ] Admin user created
- [ ] Test data prepared (optional)

### Authentication
- [ ] Login with valid credentials
- [ ] Login with invalid credentials
- [ ] Registration flow
- [ ] Logout functionality
- [ ] Session persistence
- [ ] Redirect with ?next= parameter

### Dashboard
- [ ] Unauthenticated access blocked
- [ ] Authenticated access works
- [ ] KPIs display correctly
- [ ] Tickets table loads
- [ ] Search functionality
- [ ] Status filter
- [ ] Admin panel visibility

### Ticket Management
- [ ] Create ticket
- [ ] View ticket list
- [ ] View ticket details
- [ ] Update ticket
- [ ] Add comments
- [ ] Assign ticket
- [ ] Input validation

### User Management (Admin)
- [ ] View users list
- [ ] Create user
- [ ] Update user
- [ ] Role management
- [ ] Access control (non-admin cannot access)

### Reports
- [ ] Reports page loads
- [ ] Statistics display correctly
- [ ] Access control

### Security
- [ ] Input validation
- [ ] XSS prevention
- [ ] SQL injection prevention
- [ ] Rate limiting
- [ ] Session security
- [ ] Cookie security flags

### Error Handling
- [ ] Network errors
- [ ] Server errors
- [ ] Validation errors
- [ ] Not found errors

### Cross-Browser
- [ ] Chrome
- [ ] Firefox
- [ ] Safari
- [ ] Edge

### Responsive Design
- [ ] Mobile
- [ ] Tablet
- [ ] Desktop

---

## NOTES FOR TESTERS

1. **Test Data**: Use consistent test data for reproducible results
2. **Browser Console**: Check console for errors during testing
3. **Network Tab**: Monitor API calls and responses
4. **Cookies**: Clear cookies between authentication tests if needed
5. **Environment**: Test in both dev and prod-like environments if possible
6. **Documentation**: Document any bugs or issues found during testing

---

## BUG REPORTING TEMPLATE

When reporting bugs, include:
- **Title**: Brief description
- **Steps to Reproduce**: Detailed steps
- **Expected Result**: What should happen
- **Actual Result**: What actually happened
- **Environment**: Browser, OS, user role
- **Screenshots**: If applicable
- **Console Errors**: Any JavaScript errors
- **Network Errors**: Any API errors

---

**Last Updated**: 2025-11-19
**Version**: 1.0

