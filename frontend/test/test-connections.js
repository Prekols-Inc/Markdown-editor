import dotenv from 'dotenv';
import axios from 'axios';

dotenv.config();

const STORAGE_URL = process.env.VITE_STORAGE_API_BASE_URL;
const AUTH_URL = process.env.VITE_AUTH_API_BASE_URL;

async function testConnections() {
  console.log('🧪 Testing API connections...');
  console.log('Backend URL:', STORAGE_URL);
  console.log('Auth URL:', AUTH_URL);
  
  const results = {
    storage: false,
    auth: false
  };

  try {
    console.log('🔍 Checking Storage API...');
    const storageResponse = await axios.get(`${STORAGE_URL}/health`, { timeout: 5000 });
    results.storage = storageResponse.status === 200;
    console.log('✅ Backend API is working');
  } catch (error) {
    console.log('❌ Backend API error:', error.message);
  }

  try {
    console.log('🔍 Checking Auth API...');
    const authResponse = await axios.get(`${AUTH_URL}/health`, { timeout: 5000 });
    results.auth = authResponse.status === 200;
    console.log('✅ Auth API is working');
  } catch (error) {
    console.log('❌ Auth API error:', error.message);
  }

  return results;
}

async function runTests() {
  console.log('🚀 Running API connection tests...');
  
  const results = await testConnections();
  
  console.log('\n📊 TEST RESULTS:');
  console.log(`BACKEND_API: ${results.storage ? 'PASS' : 'FAIL'}`);
  console.log(`AUTH_API: ${results.auth ? 'PASS' : 'FAIL'}`);
  
  if (results.storage && results.auth) {
    console.log('🎉 All tests passed!');
    process.exit(0);
  } else {
    console.log('💥 Some tests failed');
    process.exit(1);
  }
}

runTests().catch(error => {
  console.error('❌ Test execution error:', error);
  process.exit(1);
});