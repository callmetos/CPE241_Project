import React from 'react';

const CheckoutStepper = ({ currentStep = 1 }) => {
  const steps = [
    { id: 1, title: 'เลือกประเภท' }, 
    { id: 2, title: 'เลือกข้อมูล' }, 
    { id: 3, title: 'ข้อมูลผู้เช่า' }, 
    { id: 4, title: 'ชำระเงิน' }, 
  ];

  // Styles
  const stepperContainerStyle = {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: '30px',
    paddingBottom: '15px',
    borderBottom: '1px solid #eee',
    maxWidth: '600px',
    margin: '0 auto 30px auto', // Center the stepper
  };

  const stepStyle = (isActive, isCompleted) => ({
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    textAlign: 'center',
    color: isCompleted ? '#28a745' : (isActive ? '#0d6efd' : '#adb5bd'), // Green for completed, Blue for active, Grey for inactive
    position: 'relative', // For connector lines
    flex: 1, // Distribute space evenly
  });

  const stepNumberStyle = (isActive, isCompleted) => ({
    width: '30px',
    height: '30px',
    borderRadius: '50%',
    backgroundColor: isCompleted ? '#28a745' : (isActive ? '#0d6efd' : '#e9ecef'),
    color: isCompleted || isActive ? 'white' : '#6c757d',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
    fontWeight: 'bold',
    marginBottom: '5px',
    border: `2px solid ${isCompleted ? '#28a745' : (isActive ? '#0d6efd' : '#ced4da')}`,
    zIndex: 1, // Keep number above line
  });

  const stepTitleStyle = {
    fontSize: '0.85em',
    fontWeight: '500',
  };

  // Basic connector line (more complex styling might need pseudo-elements in CSS)
  const connectorStyle = {
      position: 'absolute',
      top: '16px', // Align with center of circle
      left: '50%',
      width: '100%',
      height: '2px',
      backgroundColor: '#e9ecef', // Default line color
      zIndex: 0, // Behind the step number
  };
   const connectorActiveStyle = {
       ...connectorStyle,
       backgroundColor: '#28a745', // Green line for completed steps
   };


  return (
    <div style={stepperContainerStyle}>
      {steps.map((step, index) => {
        const isActive = step.id === currentStep;
        const isCompleted = step.id < currentStep;
        return (
          <div key={step.id} style={stepStyle(isActive, isCompleted)}>
             {/* Connector Line (except for the first step) */}
             {index > 0 && (
                <div style={isCompleted ? connectorActiveStyle : connectorStyle}></div>
             )}
            <div style={stepNumberStyle(isActive, isCompleted)}>{step.id}</div>
            <div style={stepTitleStyle}>{step.title}</div>
          </div>
        );
      })}
    </div>
  );
};

export default CheckoutStepper;
